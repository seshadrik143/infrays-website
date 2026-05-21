// Admin portal API client. Cookie-authenticated like the customer
// portal, but on a separate cookie name (np_admin_session).

export interface AdminProfile {
  id: string;
  email: string;
  role: string;
  mfa_enrolled: boolean;
  mfa_verified: boolean;
  last_login?: string;
  created_at: string;
}

export interface AdminCustomer {
  id: string;
  email: string;
  name: string;
  company: string;
  status: string;
  stripe_customer_id?: string;
  email_verified: boolean;
  created_at: string;
}

export interface AdminSubscription {
  id: string;
  tier: string;
  status: string;
  stripe_subscription_id?: string;
  current_period_end: string;
  trial_end?: string;
  manual_offline: boolean;
  created_at: string;
}

export interface AdminDeployment {
  id: string;
  deployment_id: string;
  deployment_name: string;
  last_seen_at: string;
  last_version: string;
  flagged_for_review: boolean;
  flag_reason?: string;
}

export interface AdminEnrollmentToken {
  id: string;
  subscription_id: string;
  label?: string;
  created_at: string;
  expires_at: string;
  consumed_at?: string;
  consumed_by_deployment?: string;
}

export interface AdminLicense {
  license_id: string;
  jti: string;
  deployment_id: string;
  tier: string;
  expires_at: string;
  revoked: boolean;
  kid: string;
}

export interface CustomerDetail {
  customer: AdminCustomer;
  subscriptions: AdminSubscription[] | null;
  deployments: AdminDeployment[] | null;
  enrollment_tokens: AdminEnrollmentToken[] | null;
  licenses: AdminLicense[] | null;
}

export interface AuditEntry {
  seq: number;
  hash: string;
  event_type: string;
  customer_id?: string;
  actor: string;
  payload?: Record<string, unknown>;
  created_at: string;
}

export class AdminApiError extends Error {
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
  if (!resp.ok) throw new AdminApiError(resp.status, data?.error || resp.statusText);
  return data as T;
}

export const adminApi = {
  login: (b: { email: string; password: string }) =>
    request<{ id: string; email: string; mfa_enrolled: boolean }>("POST", "/api/admin/auth/login", b),
  me: () => request<AdminProfile>("GET", "/api/admin/auth/me"),
  logout: () => request<{ status: string }>("POST", "/api/admin/auth/logout"),
  mfaChallenge: (code: string) =>
    request<{ status: string }>("POST", "/api/admin/auth/mfa-challenge", { code }),
  mfaSetup: () =>
    request<{ secret: string; otpauth_url: string }>("POST", "/api/admin/auth/mfa-setup"),
  mfaSetupVerify: (code: string) =>
    request<{ status: string }>("POST", "/api/admin/auth/mfa-setup-verify", { code }),
  changePassword: (b: { old_password: string; new_password: string }) =>
    request<{ status: string }>("POST", "/api/admin/auth/change-password", b),

  listCustomers: (q?: string) =>
    request<{ customers: AdminCustomer[] }>("GET", `/api/admin/customers${q ? `?q=${encodeURIComponent(q)}` : ""}`),
  getCustomer: (id: string) => request<CustomerDetail>("GET", `/api/admin/customers/${encodeURIComponent(id)}`),
  updateCustomer: (id: string, patch: { status?: string; name?: string; company?: string }) =>
    request<AdminCustomer>("PATCH", `/api/admin/customers/${encodeURIComponent(id)}`, patch),
  createOfflineSubscription: (id: string, b: { tier: string; current_period_end: string; trial_end?: string }) =>
    request<AdminSubscription>("POST", `/api/admin/customers/${encodeURIComponent(id)}/subscriptions`, b),
  createEnrollmentToken: (id: string, b: { subscription_id: string; label: string; ttl_hours: number }) =>
    request<{ token: AdminEnrollmentToken; plaintext: string }>(
      "POST",
      `/api/admin/customers/${encodeURIComponent(id)}/enrollment-tokens`,
      b
    ),
  flagDeployment: (id: string, b: { flagged: boolean; reason: string }) =>
    request<AdminDeployment>("POST", `/api/admin/deployments/${encodeURIComponent(id)}/flag`, b),
  revokeLicense: (jti: string, reason: string) =>
    request<{ status: string }>("POST", `/api/admin/licenses/${encodeURIComponent(jti)}/revoke`, { reason }),
  listAudit: (limit = 100) =>
    request<{ entries: AuditEntry[] }>("GET", `/api/admin/audit?limit=${limit}`),
};
