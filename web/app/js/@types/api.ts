export interface APIResponse {
  message: string;
  errors: Record<string, string | undefined>;
  data: Record<string, unknown>;
}
