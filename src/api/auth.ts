import { apiRequest, readCookie } from "./client";

export interface AuthUser {
  id: string;
  email: string;
  display_name?: string;
  role: "user" | "admin";
  status: "pending_verification" | "active" | "suspended" | "deleted";
}

export interface RegisterRequestPayload {
  email: string;
  password: string;
  confirm_password: string;
  turnstile_token: string;
}

export interface RegisterVerifyPayload {
  email: string;
  code: string;
}

export interface LoginPayload {
  email: string;
  password: string;
  turnstile_token?: string;
}

export interface PasswordResetRequestPayload {
  email: string;
  turnstile_token: string;
}

export interface PasswordResetVerifyPayload {
  email: string;
  proof: string;
  new_password: string;
  confirm_password: string;
}

export async function requestRegisterCode(payload: RegisterRequestPayload) {
  return apiRequest<{
    status: "verification_required";
    email_hint: string;
    expires_in_seconds: number;
    resend_after_seconds: number;
  }>("/auth/register/request-code", { method: "POST", body: JSON.stringify(payload) });
}

export async function verifyRegisterCode(payload: RegisterVerifyPayload) {
  return apiRequest<{
    status: "authenticated";
    user: AuthUser;
  }>("/auth/register/verify-code", { method: "POST", body: JSON.stringify(payload) });
}

export async function login(payload: LoginPayload) {
  return apiRequest<{
    status: "authenticated";
    user: AuthUser;
  }>("/auth/login", { method: "POST", body: JSON.stringify(payload) });
}

export async function logout() {
  const csrf = readCookie("podcast_hub_csrf");
  return apiRequest<{ status: "logged_out" }>("/auth/logout", {
    method: "POST",
    headers: csrf ? { "X-CSRF-Token": csrf } : undefined
  });
}

export async function requestPasswordReset(payload: PasswordResetRequestPayload) {
  return apiRequest<{
    status: "reset_instructions_sent_if_account_exists";
    resend_after_seconds: number;
  }>("/auth/password-reset/request", { method: "POST", body: JSON.stringify(payload) });
}

export async function verifyPasswordReset(payload: PasswordResetVerifyPayload) {
  return apiRequest<{
    status: "password_reset";
    sessions_revoked: boolean;
  }>("/auth/password-reset/verify", { method: "POST", body: JSON.stringify(payload) });
}

export async function getMe() {
  return apiRequest<{ user: AuthUser }>("/auth/me");
}
