import { useParams } from "@remix-run/react";
import { useEffect, useState } from "react";
import { getFileUrl } from "~/lib/getFileUrl";
import { IAlbum } from "~/lib/types/Album";
import { ISong } from "~/lib/types/Song";
import axiosInstance from "~/utils/axiosInstance";

const AlbumDetail = () => {
  const params = useParams();

  const [album, setAlbum] = useState<IAlbum | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [unassociatedSongs, setUnassociatedSongs] = useState<ISong[]>([]);

  useEffect(() => {
    const fetchAlbum = async (id: number) => {
      try {
        const response = await axiosInstance.get(`/albums/${id}`);
        if (response.status === 200) {
          setAlbum(response.data);
        } else {
          setError("Failed to fetch album.");
        }
      } catch (err) {
        setError("An error occurred while fetching the album.");
      } finally {
        setLoading(false);
      }
    };

    const fetchUnassociatedSongs = async () => {
      try {
        const response = await axiosInstance.get("/songs/unassociated");
        if (response.status === 200) {
          setUnassociatedSongs(response.data.data);
        } else {
          setError("Failed to fetch unassociated songs.");
        }
      } catch (err) {
        setError("An error occurred while fetching unassociated songs.");
      }
    };

    fetchUnassociatedSongs();
    if (!isNaN(+(params.id || ""))) {
      fetchAlbum(+(params.id || ""));
    } else {
      setError("Invalid album ID");
      setLoading(false);
    }
  }, [params.id]);

  const assignSongToAlbum = async (songId: number) => {
    try {
      const response = await axiosInstance.get(`/albums/${params.id}/${songId}`);
      if (response.status === 200) {
        setLoading(true);
        const fetchAlbum = async (id: number) => {
          try {
            const response = await axiosInstance.get(`/albums/${id}`);
            if (response.status === 200) {
              setAlbum(response.data);
            } else {
              setError("Failed to fetch album.");
            }
          } catch (err) {
            setError("An error occurred while fetching the album.");
          } finally {
            setLoading(false);
          }
        };

        const fetchUnassociatedSongs = async () => {
          try {
            const response = await axiosInstance.get("/songs/unassociated");
            if (response.status === 200) {
              setUnassociatedSongs(response.data.data);
            } else {
              setError("Failed to fetch unassociated songs.");
            }
          } catch (err) {
            setError("An error occurred while fetching unassociated songs.");
          }
        };

        fetchUnassociatedSongs();
        fetchAlbum(+(params.id || ""));
        setLoading(false);
      } else {
        setError("Failed to assign song to album.");
      }
    } catch (err) {
      setError("An error occurred while assigning the song to the album.");
    }
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="text-red-500">{error}</div>;
  }

  return (
    <main className="p-4 w-0 grow mx-auto">
      <div className="border rounded-lg shadow-md p-6">
        <h1 className="text-2xl font-bold mb-4">{album?.Name}</h1>
        <img
          src={getFileUrl(album?.PicFilePath || "")}
          alt={album?.Name}
          className="w-full max-h-96 object-cover mb-6 rounded-md"
        />

        <h2 className="text-xl font-semibold mb-3">Songs</h2>
        {album?.Songs && album.Songs.length > 0 ? (
          <ul className="space-y-2">
            {album.Songs.map((song) => (
              <li key={song.ID} className="border rounded p-3 hover:bg-gray-800 transition">
                <p className="text-lg font-medium">{song.Name}</p>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-gray-500">No songs available.</p>
        )}

        <h2 className="text-xl font-semibold mt-6 mb-3">Unassociated Songs</h2>
        {unassociatedSongs.length > 0 ? (
          <ul className="space-y-2">
            {unassociatedSongs.map((song) => (
              <li key={song.ID} className="border rounded p-3 hover:bg-gray-800 transition flex justify-between items-center">
                <div>
                  <p className="text-lg font-medium">{song.Name}</p>
                </div>
                <button
                  className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 transition"
                  onClick={() => assignSongToAlbum(song.ID)}>
                  Assign
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-gray-500">No unassociated songs available.</p>
        )}
      </div>
    </main>
  );
};

export default AlbumDetail;
