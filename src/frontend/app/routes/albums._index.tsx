import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import axiosInstance from "~/utils/axiosInstance";
import Button from "~/components/general/Button";
import { IAlbum } from "~/lib/types/Album";
import { getFileUrl } from "~/lib/getFileUrl";
import { isAxiosError } from "axios";

const Albums: React.FC = () => {
  const navigate = useNavigate();
  const [albums, setAlbums] = useState<IAlbum[]>([]);
  const [page, setPage] = useState<number>(1);
  const [pageSize] = useState<number>(9);
  const [totalPages, setTotalPages] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(false);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [imageSearchLoading, setImageSearchLoading] = useState<boolean>(false);
  const [benchmarkTime, setBenchmarkTime] = useState<number>(0);
  const [isBenchmarking, setIsBenchmarking] = useState<boolean>(false);
  const [similarities, setSimilarities] = useState<number[]>([]);
  const [isSimilarity, setIsSimilarity] = useState<boolean>(false);

  const fetchAlbums = useCallback(
    async (page: number, search = ""): Promise<void> => {
      setIsBenchmarking(false);
      setIsSimilarity(false);
      setSimilarities([]);
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

  const handleImageSearch = async (
    event: React.ChangeEvent<HTMLInputElement>
  ): Promise<void> => {
    const file = event.target.files?.[0];
    if (file) {
      setImageSearchLoading(true);
      const formData = new FormData();
      formData.append("file", file);

      try {
        const response = await axiosInstance.post(
          "/albums/search-by-image",
          formData,
          {
            headers: {
              "Content-Type": "multipart/form-data",
            },
          }
        );

        if (response.status === 200) {
          setIsBenchmarking(true);
          setBenchmarkTime(response.data.time);
          setAlbums(response.data.data);

          const similarities = response.data.data.map(
            (item: { similarity: number }) => item.similarity
          );
          setSimilarities(similarities);

          setIsSimilarity(true);
          setTotalPages(1); // since we're not paginating the image search results
        } else {
          console.error("Unexpected response status:", response.status);
        }
      } catch (error) {
        if (isAxiosError(error) && error.response?.status === 404) {
          setAlbums([]); // Clear albums
          setTotalPages(1); // Reset pagination
          setIsSimilarity(false); // Ensure similarity mode is off
          console.error("No similar image found.");
        } else {
          console.error("Error searching by image:", error);
        }
      } finally {
        setImageSearchLoading(false);
      }
    }
  };

  useEffect(() => {
    fetchAlbums(page);
  }, [fetchAlbums, page]);

  return (
    <main className="p-4 w-0 grow overflow-y-scroll max-h-screen">
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
        <input
          type="file"
          accept=".jpg,.png,.jpeg,.webp"
          onChange={handleImageSearch}
          className="mt-4 p-2 border rounded text-white"
        />
        {isBenchmarking && (
          <p className="mt-2">
            Benchmark time: {benchmarkTime}s (checking similarity)
          </p>
        )}
      </section>

      <section className="grid grid-cols-3 gap-3 h-[400px] min-h-fit place-content-start">
        {loading || imageSearchLoading ? (
          <p className="text-center col-span-3">Loading...</p>
        ) : albums.length > 0 ? (
          albums.map((album, index) => (
            <div
              key={index}
              className="border rounded-lg cursor-pointer group hover:brightness-110 flex flex-col h-[170px] min-h-fit justify-start overflow-clip relative"
            >
              <div className="relative w-full h-[60%] overflow-hidden">
                <img
                  src={getFileUrl(album.PicFilePath)}
                  alt={album.Name}
                  className="w-full group-hover:scale-110 h-full object-cover transition-transform duration-100 ease-in"
                />
                <div className="w-full h-full absolute top-0 left-0 bg-gradient-to-t bg-black/10 from-[#202120]/100 to-[#202120]/10" />
              </div>
              <div className="p-4 h-[40%] w-full">
                <h2 className="font-semibold line-clamp-1 w-full">
                  {album.Name}
                </h2>
                <div className="flex justify-between items-center">
                  <p className="text-xs text-gray-200">ID: {album.ID}</p>
                  {isSimilarity && (
                    <p className="text-xs text-gray-200">
                      Similarity: {Math.round(similarities[index] * 100)}%
                    </p>
                  )}
                </div>
              </div>
            </div>
          ))
        ) : (
          <p className="text-center col-span-3">No albums found.</p>
        )}
      </section>

      <section className="flex justify-between items-center mt-6">
        <Button
          onClick={() => handlePageChange(page - 1)}
          disabled={page === 1}
        >
          Previous
        </Button>
        <span>
          Page {page} of {totalPages}
        </span>
        <Button
          onClick={() => handlePageChange(page + 1)}
          disabled={page === totalPages}
        >
          Next
        </Button>
      </section>
    </main>
  );
};

export default Albums;
