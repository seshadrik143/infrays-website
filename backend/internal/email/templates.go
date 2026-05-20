// Templates for each transactional email type.
//
// Kept as Go string constants rather than separate files so the
// package is self-contained — the issuer binary embeds them via
// normal Go compilation, no //go:embed needed, no asset bundling.
// Adding a sixth template = add a const + a Render function.
//
// All templates use html/template for HTML and text/template for
// text. The HTML versions include light inline CSS — Gmail / Outlook
// strip <style> blocks, so inline is the only portable approach.
//
// The Render* functions return ready-to-send Message structs;
// callers don't need to know about template internals.
package email

import (
	"bytes"
	htmltpl "html/template"
	"strings"
	texttpl "text/template"
)

// ── Data structs each template renders against ────────────────────

// WelcomeData is what RenderWelcome reads.
type WelcomeData struct {
	CustomerName string
	AppURL       string
	DocsURL      string
}

// TrialExpiringData covers the 30/7/1-day reminder cadence. The
// caller picks which Day to send for; the template renders the
// urgency text accordingly.
type TrialExpiringData struct {
	CustomerName string
	DaysLeft     int
	AppURL       string
	UpgradeURL   string // typically AppURL + "/billing"
}

// PaymentFailedData fires on invoice.payment_failed. Stripe's own
// dunning emails go too — this is the version with our branding +
// pointer to the customer portal.
type PaymentFailedData struct {
	CustomerName  string
	AmountDueUSD  string // pre-formatted "$X.YZ"
	AppURL        string
	UpdateCardURL string // app.infrays.org/billing
}

// SubscriptionCanceledData covers the customer.subscription.deleted
// event. License stays active until grace period elapses.
type SubscriptionCanceledData struct {
	CustomerName string
	AccessEndsOn string // formatted date "May 20, 2027"
	AppURL       string
	ContactURL   string // mailto: link
}

// EnrollmentTokenCreatedData is sent when sales / customer mints a
// new enrollment token. Includes the plaintext token (only place it
// appears in transit — afterwards only the hash is stored).
type EnrollmentTokenCreatedData struct {
	CustomerName    string
	EnrollmentToken string
	Label           string // optional "prod-cluster-east"
	ExpiresIn       string // "24 hours"
	DocsURL         string
}

// ── Renderers ────────────────────────────────────────────────────

func RenderWelcome(to string, d WelcomeData) (Message, error) {
	return renderBoth(to, "Welcome to infraYS", "welcome", welcomeHTML, welcomeText, d)
}

func RenderTrialExpiring(to string, d TrialExpiringData) (Message, error) {
	subj := "Your NodePulse trial expires in " + intToStr(d.DaysLeft) + " day"
	if d.DaysLeft != 1 {
		subj += "s"
	}
	return renderBoth(to, subj, "trial_expiring", trialExpiringHTML, trialExpiringText, d)
}

func RenderPaymentFailed(to string, d PaymentFailedData) (Message, error) {
	return renderBoth(to, "Payment failed for your NodePulse subscription", "payment_failed", paymentFailedHTML, paymentFailedText, d)
}

func RenderSubscriptionCanceled(to string, d SubscriptionCanceledData) (Message, error) {
	return renderBoth(to, "Your NodePulse subscription was canceled", "subscription_canceled", subscriptionCanceledHTML, subscriptionCanceledText, d)
}

func RenderEnrollmentTokenCreated(to string, d EnrollmentTokenCreatedData) (Message, error) {
	return renderBoth(to, "Your NodePulse enrollment token", "enrollment_token_created", enrollmentTokenHTML, enrollmentTokenText, d)
}

// renderBoth executes both HTML and text templates and packages the
// result as a Message.
func renderBoth(to, subject, msgType, htmlSrc, textSrc string, data any) (Message, error) {
	var htmlOut, textOut bytes.Buffer

	ht, err := htmltpl.New(msgType + "_html").Parse(htmlSrc)
	if err != nil {
		return Message{}, err
	}
	if err := ht.Execute(&htmlOut, data); err != nil {
		return Message{}, err
	}

	tt, err := texttpl.New(msgType + "_text").Parse(textSrc)
	if err != nil {
		return Message{}, err
	}
	if err := tt.Execute(&textOut, data); err != nil {
		return Message{}, err
	}

	return Message{
		To:          to,
		Subject:     subject,
		HTMLBody:    htmlOut.String(),
		TextBody:    textOut.String(),
		MessageType: msgType,
	}, nil
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToStr(-n)
	}
	var out strings.Builder
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	out.Write(digits)
	return out.String()
}

// ── Template bodies ──────────────────────────────────────────────
//
// Kept tight. Branded but not flashy — these go to operators and
// platform engineers, not consumer mailing lists. Inline CSS only.

const baseStyle = `font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; color: #1f2937; line-height: 1.5; max-width: 560px;`

const welcomeHTML = `<!DOCTYPE html><html><body style="` + baseStyle + `">
<h2 style="color: #15b370;">Welcome to infraYS, {{.CustomerName}}</h2>
<p>Your NodePulse account is active. The next step is to install the agent and connect it to your local NodePulse server.</p>
<ul>
  <li><a href="{{.DocsURL}}">Quick-start guide</a></li>
  <li><a href="{{.AppURL}}/deployments">Manage your deployments</a></li>
  <li><a href="{{.AppURL}}/enrollment-tokens">Generate enrollment tokens</a></li>
</ul>
<p>Questions? Just reply to this email — it reaches our support team.</p>
<p style="color: #6b7280; font-size: 13px; margin-top: 32px;">— infraYS</p>
</body></html>`

const welcomeText = `Welcome to infraYS, {{.CustomerName}}

Your NodePulse account is active. The next step is to install the agent and connect it to your local NodePulse server.

  Quick-start guide:        {{.DocsURL}}
  Manage your deployments:  {{.AppURL}}/deployments
  Generate enrollment tokens: {{.AppURL}}/enrollment-tokens

Questions? Reply to this email — it reaches our support team.

— infraYS
`

const trialExpiringHTML = `<!DOCTYPE html><html><body style="` + baseStyle + `">
<h2 style="color: #f59e0b;">Your NodePulse trial expires in {{.DaysLeft}} day{{if ne .DaysLeft 1}}s{{end}}</h2>
<p>Hi {{.CustomerName}},</p>
<p>Your 15-day NodePulse trial wraps up soon. To keep monitoring without interruption, upgrade to a paid plan from your billing page.</p>
<p><a href="{{.UpgradeURL}}" style="display: inline-block; background: #15b370; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none;">Upgrade now</a></p>
<p>After expiry, your local NodePulse server enters a 30-day read-only grace period. Core monitoring keeps working; Enterprise features (SSO, audit log, advanced alerts) are restricted until renewal.</p>
<p style="color: #6b7280; font-size: 13px; margin-top: 32px;">— infraYS</p>
</body></html>`

const trialExpiringText = `Your NodePulse trial expires in {{.DaysLeft}} day{{if ne .DaysLeft 1}}s{{end}}

Hi {{.CustomerName}},

Your 15-day NodePulse trial wraps up soon. To keep monitoring without interruption, upgrade to a paid plan:

  {{.UpgradeURL}}

After expiry, your local NodePulse server enters a 30-day read-only grace period. Core monitoring keeps working; Enterprise features are restricted until renewal.

— infraYS
`

const paymentFailedHTML = `<!DOCTYPE html><html><body style="` + baseStyle + `">
<h2 style="color: #ef4444;">Payment failed</h2>
<p>Hi {{.CustomerName}},</p>
<p>We were unable to process your last NodePulse subscription payment of <strong>{{.AmountDueUSD}}</strong>.</p>
<p>Most commonly this is because the card on file expired or hit a temporary auth check. Update your payment method and we'll retry automatically.</p>
<p><a href="{{.UpdateCardURL}}" style="display: inline-block; background: #15b370; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none;">Update payment method</a></p>
<p>Stripe will retry payment over the next few days. After several failed attempts your subscription will be canceled and your NodePulse license will enter a 90-day grace period before Enterprise features go offline. Core monitoring stays operational throughout.</p>
<p style="color: #6b7280; font-size: 13px; margin-top: 32px;">— infraYS</p>
</body></html>`

const paymentFailedText = `Payment failed

Hi {{.CustomerName}},

We were unable to process your last NodePulse subscription payment of {{.AmountDueUSD}}.

Update your payment method:
  {{.UpdateCardURL}}

Stripe will retry over the next few days. After several failed attempts your subscription will be canceled and your license enters a 90-day grace period. Core monitoring stays operational throughout.

— infraYS
`

const subscriptionCanceledHTML = `<!DOCTYPE html><html><body style="` + baseStyle + `">
<h2>Your NodePulse subscription was canceled</h2>
<p>Hi {{.CustomerName}},</p>
<p>This confirms your NodePulse subscription was canceled. Your existing license remains valid until <strong>{{.AccessEndsOn}}</strong>; after that, a 90-day read-only grace window kicks in before Enterprise features go offline.</p>
<p>Core monitoring (metrics, logs, traces, basic alerts) will continue running on your local NodePulse server indefinitely at Community-tier limits.</p>
<p>If you canceled by mistake, we can reinstate your subscription — just reply to this email or use <a href="{{.ContactURL}}">our support form</a>.</p>
<p style="color: #6b7280; font-size: 13px; margin-top: 32px;">— infraYS</p>
</body></html>`

const subscriptionCanceledText = `Your NodePulse subscription was canceled

Hi {{.CustomerName}},

This confirms your NodePulse subscription was canceled. Your existing license remains valid until {{.AccessEndsOn}}; after that, a 90-day read-only grace window kicks in before Enterprise features go offline.

Core monitoring (metrics, logs, traces, basic alerts) will continue running on your local NodePulse server indefinitely at Community-tier limits.

If this was a mistake, reply to this email or use {{.ContactURL}}.

— infraYS
`

const enrollmentTokenHTML = `<!DOCTYPE html><html><body style="` + baseStyle + `">
<h2>Your NodePulse enrollment token</h2>
<p>Hi {{.CustomerName}},</p>
<p>An enrollment token has been generated for your account{{if .Label}} (label: <code>{{.Label}}</code>){{end}}. Paste it into your NodePulse server config:</p>
<pre style="background: #1f2937; color: #d1d5db; padding: 16px; border-radius: 6px; overflow-x: auto;"># /etc/nodepulse/server.yaml
license:
  enrollment_token: "{{.EnrollmentToken}}"</pre>
<p>Or set it as an environment variable to keep it out of config:</p>
<pre style="background: #1f2937; color: #d1d5db; padding: 16px; border-radius: 6px; overflow-x: auto;">export NP_LICENSE_ENROLLMENT_TOKEN="{{.EnrollmentToken}}"</pre>
<p><strong>This token expires in {{.ExpiresIn}}.</strong> It's single-use per deployment — once your NodePulse server exchanges it for a signed license, the token is sealed. You can regenerate it from your portal if you need a fresh one.</p>
<p>Full installation guide: <a href="{{.DocsURL}}">{{.DocsURL}}</a></p>
<p style="color: #6b7280; font-size: 13px; margin-top: 32px;">— infraYS</p>
</body></html>`

const enrollmentTokenText = `Your NodePulse enrollment token

Hi {{.CustomerName}},

An enrollment token has been generated for your account{{if .Label}} (label: {{.Label}}){{end}}. Paste it into your NodePulse server config:

  # /etc/nodepulse/server.yaml
  license:
    enrollment_token: "{{.EnrollmentToken}}"

Or set as env var:

  export NP_LICENSE_ENROLLMENT_TOKEN="{{.EnrollmentToken}}"

This token expires in {{.ExpiresIn}}. Single-use per deployment — once exchanged for a signed license, it's sealed. Regenerate from your portal if needed.

Full guide: {{.DocsURL}}

— infraYS
`
