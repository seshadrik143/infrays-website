// Minimal fetch wrapper for the portal. All endpoints live under
// /api/portal/* and use cookie-based auth — no token plumbing on the
// frontend.

export interface CustomerProfile {
  id: string;
  email: string;
  name: string;
  company: string;
  status: string;
  email_verified: boolean;
  created_at: string;
}

export interface Subscription {
  id: string;
  tier: string;
  status: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at?: string;
  canceled_at?: string;
  trial_end?: string;
  stripe_subscription_id?: string;
  manual_offline: boolean;
  created_at: string;
}

export interface Deployment {
  id: string;
  deployment_id: string;
  deployment_name: string;
  first_seen_at: string;
  last_seen_at: string;
  last_version: string;
  flagged_for_review: boolean;
  flag_reason?: string;
}

export interface EnrollmentToken {
  id: string;
  subscription_id: string;
  label?: string;
  created_at: string;
  expires_at: string;
  consumed_at?: string;
  consumed_by_deployment?: string;
}

export interface License {
  license_id: string;
  jti: string;
  deployment_id: string;
  tier: string;
  issued_at: string;
  expires_at: string;
  grace_until: string;
  revoked: boolean;
  kid: string;
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, msg: string) {
    super(msg);
    this.status = status;
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = {
    method,
    credentials: "include",
    headers: { "Content-Type": "application/json" },
  };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const resp = await fetch(path, opts);
  let data: any = null;
  const text = await resp.text();
  try { data = text ? JSON.parse(text) : null; } catch { data = { raw: text }; }
  if (!resp.ok) {
    throw new ApiError(resp.status, data?.error || resp.statusText);
  }
  return data as T;
}

export const api = {
  signup: (b: { email: string; password: string; name?: string; company?: string }) =>
    request<{ status: string }>("POST", "/api/portal/auth/signup", b),
  login: (b: { email: string; password: string }) =>
    request<CustomerProfile>("POST", "/api/portal/auth/login", b),
  logout: () => request<{ status: string }>("POST", "/api/portal/auth/logout"),
  me: () => request<CustomerProfile>("GET", "/api/portal/auth/me"),
  verifyEmail: (token: string) =>
    request<{ status: string }>("POST", "/api/portal/auth/verify-email", { token }),
  resendVerification: () =>
    request<{ status: string }>("POST", "/api/portal/auth/resend-verification"),
  changePassword: (b: { old_password: string; new_password: string }) =>
    request<{ status: string }>("POST", "/api/portal/auth/change-password", b),
  requestPasswordReset: (email: string) =>
    request<{ status: string }>("POST", "/api/portal/auth/request-password-reset", { email }),
  resetPassword: (b: { token: string; new_password: string }) =>
    request<{ status: string }>("POST", "/api/portal/auth/reset-password", b),

  listSubscriptions: () =>
    request<{ subscriptions: Subscription[] }>("GET", "/api/portal/subscriptions"),
  listDeployments: () =>
    request<{ deployments: Deployment[] }>("GET", "/api/portal/deployments"),
  listEnrollmentTokens: () =>
    request<{ tokens: EnrollmentToken[] }>("GET", "/api/portal/enrollment-tokens"),
  createEnrollmentToken: (b: { subscription_id: string; label: string; ttl_hours: number }) =>
    request<EnrollmentToken & { plaintext: string }>(
      "POST",
      "/api/portal/enrollment-tokens",
      b
    ),
  revokeEnrollmentToken: (token_id: string) =>
    request<{ status: string }>("POST", "/api/portal/enrollment-tokens/revoke", { token_id }),
  listLicenses: () => request<{ licenses: License[] }>("GET", "/api/portal/licenses"),
  offlineLicenseURL: (license_id: string) =>
    `/api/portal/offline-license?license_id=${encodeURIComponent(license_id)}`,
  updateAccount: (b: { name?: string; company?: string }) =>
    request<CustomerProfile>("PATCH", "/api/portal/account", b),
  billingPortalURL: () =>
    request<{ url: string }>("POST", "/api/portal/billing-portal-url"),
};
