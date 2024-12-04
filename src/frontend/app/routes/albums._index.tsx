import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import axiosInstance from "~/utils/axiosInstance";
import Button from "~/components/general/Button";
import { IAlbum } from "~/lib/types/Album";

const Albums: React.FC = () => {
  const navigate = useNavigate();
  const [albums, setAlbums] = useState<IAlbum[]>([]);
  const [page, setPage] = useState<number>(1);
  const [pageSize] = useState<number>(9);
  const [totalPages, setTotalPages] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(false);
  const [searchQuery, setSearchQuery] = useState<string>("");

  const fetchAlbums = useCallback(
    async (page: number, search = ""): Promise<void> => {
      setLoading(true);
      try {
        const response = await axiosInstance.get("/albums", {
          params: { page, page_size: pageSize, search },
        });

        if (response.status === 200) {
          setAlbums(response.data.data);
          setTotalPages(Math.ceil(response.data.totalItems / pageSize));
        } else {
          console.error("Failed to fetch albums.");
        }
      } catch (error) {
        console.error("Error fetching albums:", error);
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
    fetchAlbums(1, searchQuery);
    setPage(1);
  };

  const handlePageChange = (newPage: number): void => {
    if (newPage > 0 && newPage <= totalPages) {
      setPage(newPage);
      fetchAlbums(newPage, searchQuery);
    }
  };

  const handleNavigateToUpload = (): void => {
    navigate("/albums/upload");
  };

  useEffect(() => {
    fetchAlbums(page);
  }, [fetchAlbums, page]);

  return (
    <main className="p-4 w-0 grow">
      <section className="flex w-full justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Albums</h1>
        <Button onClick={handleNavigateToUpload}>+ Upload Album</Button>
      </section>

      <section className="mb-6">
        <div className="flex items-center space-x-2">
          <input
            type="text"
            placeholder="Search albums..."
            value={searchQuery}
            onChange={handleSearch}
            className="p-2 border rounded flex-grow text-black"
          />
          <Button onClick={handleSearchSubmit}>Search</Button>
        </div>
      </section>

      <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 h-[400px] place-content-start">
        {loading ? (
          <p className="text-center col-span-3">Loading albums...</p>
        ) : albums.length > 0 ? (
          albums.map((album, index) => (
            <div key={index} className="border p-4 rounded-lg shadow-md flex flex-col h-[130px] justify-start overflow-clip">
              <h2 className="font-semibold line-clamp-1 w-full">{album.Name}</h2>
              <p className="text-sm text-gray-600">ID: {album.ID}</p>
            </div>
          ))
        ) : (
          <p className="text-center col-span-3">No albums found.</p>
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

export default Albums;
