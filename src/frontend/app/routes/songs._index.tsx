import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axiosInstance from "~/utils/axiosInstance";
import Button from "~/components/general/Button";
import { ISong } from "~/lib/types/Song";

const Songs: React.FC = () => {
  const navigate = useNavigate();
  const [songs, setSongs] = useState<ISong[]>([]);
  const [page, setPage] = useState<number>(1);
  const [pageSize] = useState<number>(9); // Fixed page size
  const [totalPages, setTotalPages] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(false);
  const [searchQuery, setSearchQuery] = useState<string>("");

  const fetchSongs = async (page: number, search = ""): Promise<void> => {
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
  };

  const handleSearch = (event: React.ChangeEvent<HTMLInputElement>): void => {
    setSearchQuery(event.target.value);
  };

  const handleSearchSubmit = (): void => {
    fetchSongs(1, searchQuery); // Fetch songs with the search query
    setPage(1); // Reset to the first page
  };

  const handlePageChange = (newPage: number): void => {
    if (newPage > 0 && newPage <= totalPages) {
      setPage(newPage);
      fetchSongs(newPage, searchQuery);
    }
  };

  useEffect(() => {
    fetchSongs(page);
  }, [page]);

  const handleNavigateToUpload = (): void => {
    navigate("/songs/upload");
  };

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
      </section>

      <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 h-[400px] place-content-start">
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
        <Button
          onClick={() => handlePageChange(page - 1)}
          disabled={page === 1} // Disable when on the first page
          className={`${page === 1 && "bg-gray-500 cursor-not-allowed hover:bg-gray-500 hover:shadow-none"}`}>
          Previous
        </Button>
        <span>
          Page {page} of {totalPages}
        </span>
        <Button
          onClick={() => handlePageChange(page + 1)}
          disabled={page === totalPages} // Disable when on the last page
          className={`${page === totalPages && "bg-gray-500 cursor-not-allowed hover:bg-gray-500 hover:shadow-none"}`}>
          Next
        </Button>
      </section>
    </main>
  );
};

export default Songs;
