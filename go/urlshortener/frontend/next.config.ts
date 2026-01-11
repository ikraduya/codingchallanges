import type { NextConfig } from "next";

const requiredEnv = ['BACKEND_URL']
for (const key of requiredEnv) {
  const val = process.env[key];
  if (!val || val.trim() === '') {
    throw new Error(`Missing required environment variable: ${key}`);
  }
}

const nextConfig: NextConfig = {
  reactStrictMode: true,
};

export default nextConfig;
