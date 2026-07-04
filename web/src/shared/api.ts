const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";

export interface HealthResponse {
  status: string;
  env: string;
  time: string;
}

export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch(`${API_BASE_URL}/health`, {
    credentials: "include"
  });

  if (!response.ok) {
    throw new Error(`Health check failed: ${response.status}`);
  }

  return response.json() as Promise<HealthResponse>;
}

