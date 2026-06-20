/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["@monti/shared"],
  experimental: {
    typedRoutes: false,
  },
};

export default nextConfig;
