import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/upload.tsx"),
  route("upload", "upload/upload.tsx"),
] satisfies RouteConfig;
