import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import axiosInstance from "~/utils/axiosInstance";
import Button from "~/components/general/Button";
import { ISong } from "~/lib/types/Song";

const Songs: React.FC = () => {
  const navigate = useNavigate();
  const [songs, setSongs] = useState<ISong[]>([]);
  const [page, setPage] = useState<number>(1);
  const [pageSize] = useState<number>(9);
  const [totalPages, setTotalPages] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(false);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [recording, setRecording] = useState<boolean>(false);
  const [countdown, setCountdown] = useState<number>(5);

  let mediaRecorder: MediaRecorder | null = null;

  const fetchSongs = useCallback(
    async (page: number, search = ""): Promise<void> => {
      setLoading(true);
      try {
        const response = await axiosInstance.get("/songs", {
          params: { page, page_size: pageSize, search },
        });

        if (response.status === 200) {
          setSongs(response.data.data);
          setTotalPages(Math.ceil(response.data.totalItems / pageSize));
        } else {
          console.error("Failed to fetch songs.");
        }
      } catch (error) {
        console.error("Error fetching songs:", error);
      } finally {
        setLoading(false);
      }
    },
    [pageSize]
  );

  const handleSearch = (event: React.ChangeEvent<HTMLInputElement>): void => {
    setSearchQuery(event.target.value);
  };

  const handleSearchSubmit = (): void => {
    fetchSongs(1, searchQuery);
    setPage(1);
  };

  const handlePageChange = (newPage: number): void => {
    if (newPage > 0 && newPage <= totalPages) {
      setPage(newPage);
      fetchSongs(newPage, searchQuery);
    }
  };

  const handleNavigateToUpload = (): void => {
    navigate("/songs/upload");
  };

  const startRecording = () => {
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      alert("Microphone not supported in your browser.");
      return;
    }

    setRecording(true);
    setCountdown(5);
    navigator.mediaDevices
      .getUserMedia({ audio: true })
      .then((stream) => {
        mediaRecorder = new MediaRecorder(stream);
        const audioChunks: Blob[] = [];

        mediaRecorder.ondataavailable = (event) => {
          audioChunks.push(event.data);
        };

        mediaRecorder.onstop = () => {
          const audioBlob = new Blob(audioChunks, { type: "audio/wav" });
          searchByHumming(audioBlob, "humming.wav");
        };

        mediaRecorder.start();

        const countdownInterval = setInterval(() => {
          setCountdown((prev) => {
            if (prev > 1) {
              return prev - 1;
            } else {
              clearInterval(countdownInterval);
              stopRecording();
              return 0;
            }
          });
        }, 1000);
      })
      .catch((error) => {
        console.error("Error accessing microphone:", error);
        setRecording(false);
      });
  };

  const stopRecording = () => {
    if (mediaRecorder && mediaRecorder.state !== "inactive") {
      mediaRecorder.stop();
    }
    setRecording(false);
    setCountdown(0);
  };

  const handleAudioUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      searchByHumming(file, file.name);
    }
  };

  const searchByHumming = async (audioBlob: Blob, filename: string) => {
    setLoading(true);
    try {
      const formData = new FormData();
      formData.append("file", audioBlob, filename); // Ensure filename is passed here

      const response = await axiosInstance.post("/songs/search-by-audio", formData);

      if (response.status === 200) {
        if (response.data.data.length > 0) {
          setSongs(response.data.data);
          setTotalPages(1); // Assuming no pagination for search results
        } else {
          setSongs([]); // No songs found
          setTotalPages(1);
        }
      } else {
        console.error("Failed to search by audio.");
      }
    } catch (error) {
      setSongs([]); // No songs found
      console.error("Error searching by audio:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSongs(page);
  }, [fetchSongs, page]);

  return (
    <main className="p-4 w-0 grow">
      <section className="flex w-full justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Songs</h1>
        <Button onClick={handleNavigateToUpload}>+ Upload Song</Button>
      </section>

      <section className="mb-6">
        <div className="flex items-center space-x-2">
          <input
            type="text"
            placeholder="Search songs..."
            value={searchQuery}
            onChange={handleSearch}
            className="p-2 border rounded flex-grow text-black"
          />
          <Button onClick={handleSearchSubmit}>Search</Button>
        </div>
        <div className="flex items-center space-x-2 mt-4">
          <Button onClick={startRecording} disabled={recording}>
            {recording ? `Recording... (${countdown}s)` : "Search by Humming"}
          </Button>
          {recording && (
            <Button onClick={stopRecording} className="bg-red-500">
              Stop
            </Button>
          )}
          <input type="file" accept="audio/*" onChange={handleAudioUpload} className="hidden" id="audio-upload" />
          <label htmlFor="audio-upload" className="cursor-pointer px-4 py-2 bg-blue-500 text-white rounded">
            Upload Audio File
          </label>
        </div>
      </section>

      <section className="grid grid-cols-3 gap-3 h-[400px] place-content-start">
        {loading ? (
          <p className="text-center col-span-3">Loading songs...</p>
        ) : songs.length > 0 ? (
          songs.map((song, index) => (
            <div key={index} className="border p-4 rounded-lg shadow-md flex flex-col h-[130px] justify-start overflow-clip">
              <h2 className="font-semibold line-clamp-1 w-full">{song.Name}</h2>
              <p className="text-sm text-gray-600">ID: {song.ID}</p>
            </div>
          ))
        ) : (
          <p className="text-center col-span-3">No songs found.</p>
        )}
      </section>

      <section className="flex justify-between items-center mt-6">
        <Button onClick={() => handlePageChange(page - 1)} disabled={page === 1}>
          Previous
        </Button>
        <span>
          Page {page} of {totalPages}
        </span>
        <Button onClick={() => handlePageChange(page + 1)} disabled={page === totalPages}>
          Next
        </Button>
      </section>
    </main>
  );
};

export default Songs;
