import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import axiosInstance from "~/utils/axiosInstance";
import Button from "~/components/general/Button";
import { ISong } from "~/lib/types/Song";
import { FaCompactDisc } from "react-icons/fa";
import { getFileUrl } from "~/lib/getFileUrl";
import Recording from "~/components/general/Recording"; // Import the new component

const Songs: React.FC = () => {
  const navigate = useNavigate();
  const [songs, setSongs] = useState<ISong[]>([]);
  const [page, setPage] = useState<number>(1);
  const [pageSize] = useState<number>(9);
  const [totalPages, setTotalPages] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(false);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [debounceTimeout, setDebounceTimeout] = useState<NodeJS.Timeout | null>(null);
  const [isSearchResult, setIsSearchResult] = useState<boolean>(false);
  const [benchmarkTime, setBenmarkTime] = useState<number>(0);

  const fetchSongs = useCallback(
    async (page: number, search = ""): Promise<void> => {
      setIsSearchResult(false);
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

  const handleSearch = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>): void => {
      setSearchQuery(event.target.value);

      // Clear previous timeout
      if (debounceTimeout) {
        clearTimeout(debounceTimeout);
      }

      // Set new timeout for debounce
      const timeoutId = setTimeout(() => {
        fetchSongs(1, event.target.value); // Fetch songs after 1 second delay
      }, 1000);

      setDebounceTimeout(timeoutId);
    },
    [debounceTimeout, fetchSongs]
  );

  const handleSearchSubmit = useCallback((): void => {
    fetchSongs(1, searchQuery);
    setPage(1);
  }, [fetchSongs, searchQuery]);

  const handlePageChange = useCallback(
    (newPage: number): void => {
      if (newPage > 0 && newPage <= totalPages) {
        setPage(newPage);
        fetchSongs(newPage, searchQuery);
      }
    },
    [fetchSongs, searchQuery, totalPages]
  );

  const handleNavigateToUpload = useCallback((): void => {
    navigate("/songs/upload");
  }, [navigate]);

  const searchByHumming = useCallback(async (audioBlob: Blob, filename: string) => {
    try {
      setIsSearchResult(true);
      const formData = new FormData();
      formData.append("file", audioBlob, filename);

      const response = await axiosInstance.post("/songs/search-by-audio", formData);
      if (response.status === 200 && response.data.data.length > 0) {
        setSongs(response.data.data);
        setTotalPages(1); // Adjust pagination
        setBenmarkTime(response.data.time);
      } else {
        console.warn("No matching songs found.");
      }
    } catch (error) {
      setSongs([]);
    }
  }, []);

  const handleAudioUpload = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      if (file) {
        searchByHumming(file, file.name);
      }
    },
    [searchByHumming]
  );

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
        <div className="flex py-2 gap-2">
          <input type="file" accept="audio/*" onChange={handleAudioUpload} className="hidden" id="audio-upload" />
          <label htmlFor="audio-upload" className="cursor-pointer px-4 py-2 bg-blue-500 text-white rounded">
            Upload Audio File
          </label>
          <Recording searchByHumming={searchByHumming} />
        </div>
      </section>

      {isSearchResult && <p>Search took {benchmarkTime}ms</p>}

      <section className="grid grid-cols-3 gap-3 h-[400px] place-content-start">
        {loading ? (
          <p className="text-center col-span-3">Loading songs...</p>
        ) : songs.length > 0 ? (
          songs.map((song, index) => (
            <a
              key={index}
              className="border rounded-lg shadow-md flex flex-row h-[130px] justify-start overflow-clip items-center hover:brightness-110 bg-[#212121] cursor-pointer"
              href={getFileUrl(song.AudioFilePath)}
              target="_blank"
              rel="noreferrer">
              <div className="w-auto h-full aspect-square p-4">
                <FaCompactDisc className="w-auto h-full aspect-square" />
              </div>
              <div className="flex w-full flex-col p-4">
                <h2 className="font-semibold line-clamp-1 w-full">{song.Name}</h2>
                <p className="text-sm text-gray-600">ID: {song.ID}</p>
                {isSearchResult && <p className="text-xs">Similarity {(+(song.SimilarityScore || 0) * 100).toFixed(2)}%</p>}
              </div>
            </a>
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
