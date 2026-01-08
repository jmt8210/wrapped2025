import { useActionData } from "react-router";
import type { Route } from "../routes/+types/upload";
// import Chart from "~/chart";
import { SongTable } from "~/SongTable";

export type TopSong = {
  Name: string;
  Artist: string;
  Skips: number;
  Plays: number;
};

export type MinutesPerDay = {
  Date: string;
  Count: number;
};

export type StatsResponse = {
  MostSkipped: TopSong;
  TopSongs: TopSong[];
  TopArtists: TopSong[];
  MinsPerDay: MinutesPerDay[];
};

export async function action({
  request,
}: Route.ActionArgs): Promise<StatsResponse> {
  const formData = await request.formData();

  const stats = await fetch("http://localhost:29228/get-stats", {
    method: "POST",
    body: formData,
  }).then((res) => res.json());

  return stats;
}

export default function Upload() {
  const actionData = useActionData<StatsResponse>();
  console.log(actionData)
  return actionData ? (
    <div
      className={`flex flex-col items-center justify-center text-white bg-[#041c0d] min-h-screen min-w-screen gap-10`}
    >
      <div>
        Most Skipped: {actionData.MostSkipped.Name},{" "}
        {actionData.MostSkipped.Skips} times
      </div>
      <div className="flex flex-row flex-wrap gap-x-20 gap-y-4 mx-4">
        <div className="flex flex-col gap-4">
          <div>50 Top Songs</div>
          <SongTable songs={actionData.TopSongs} />
        </div>
        <div className="flex flex-col gap-4">
          <div>50 Top Artists</div>
          <SongTable songs={actionData.TopArtists} />
        </div>
      </div>
      {/* <Chart data={actionData.MinsPerDay} width={300} height={300} /> */}
    </div>
  ) : (
    <div className="flex justify-center items-center bg-green-300 min-h-screen min-w-screen">
      <form method="post" encType="multipart/form-data" action="/upload">
        <input type="file" name="file" />
        <button>Submit</button>
      </form>
    </div>
  );
}
