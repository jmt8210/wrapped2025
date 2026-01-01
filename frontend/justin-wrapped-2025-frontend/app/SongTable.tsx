import type { TopSong } from "./upload/upload";

export const SongTable = ({ songs }: { songs: TopSong[] }) => {
  return (
    <table>
      <thead>
        <tr className="text-left">
          <th className="pr-2">Name</th>
          <th className="pr-2">Plays</th>
          <th className="pr-2">Skips</th>
        </tr>
      </thead>
      <tbody>
        {songs.map((song: TopSong) => (
          <tr className="odd:bg-green-800 even:bg-green-900">
            <td>{song.Name}</td>
            <td>{song.Skips}</td>
            <td>{song.Plays}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};
