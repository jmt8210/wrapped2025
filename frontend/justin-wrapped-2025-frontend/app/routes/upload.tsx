import Upload from "~/upload/upload";
import type { Route } from "./+types/upload";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Justin Spotify Wrapped 2025" },
    { name: "description", content: "See your 2025 stats for spotify" },
  ];
}

export default function Home() {
  return (
    <Upload
      actionData={undefined}
      params={[]}
      matches={[
        {
          id: "root",
          params: {},
          pathname: "",
          data: undefined,
          loaderData: undefined,
          handle: [],
        },
        {
          id: "routes/upload",
          params: {},
          pathname: "",
          data: undefined,
          loaderData: undefined,
          handle: [],
        },
      ]}
      loaderData={undefined}
    />
  );
}
