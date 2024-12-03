import { useRef, ChangeEvent, useState } from "react";
import Button from "~/components/general/Button";
import axiosInstance from "~/utils/axiosInstance";

const Songs: React.FC = () => {
  const zipInputRef = useRef<HTMLInputElement | null>(null);
  const songInputRef = useRef<HTMLInputElement | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const handleZipUpload = (event: ChangeEvent<HTMLInputElement>): void => {
    const file = event.target.files?.[0];
    if (file) {
      uploadFile(file);
    }
  };

  const handleSongUpload = (event: ChangeEvent<HTMLInputElement>): void => {
    const file = event.target.files?.[0];
    if (file) {
      uploadFile(file);
    }
  };

  const uploadFile = async (file: File): Promise<void> => {
    const formData = new FormData();
    formData.append("file", file);

    setLoading(true); // Start loading

    try {
      const response = await axiosInstance.post("/songs/upload", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (response.status === 200) {
        console.log(response.data);
      } else {
        console.error("Upload failed");
      }
    } catch (error) {
      console.error("Error uploading file:", error);
    } finally {
      setLoading(false); // Stop loading after the upload finishes
    }
  };

  return (
    <main className="p-4 w-0 grow">
      <section className="flex w-full justify-between">
        <h1>Songs</h1>
        <div className="flex gap-2">
          <Button onClick={() => zipInputRef.current?.click()}>+ zip</Button>
          <Button onClick={() => songInputRef.current?.click()}>+ song</Button>
        </div>
      </section>

      {/* Loading Indicator */}
      {loading && <div className="loading-spinner">Uploading...</div>}

      <input type="file" accept=".zip" ref={zipInputRef} onChange={handleZipUpload} className="hidden" />
      <input type="file" accept=".mp3,.wav,.mid" ref={songInputRef} onChange={handleSongUpload} className="hidden" />
    </main>
  );
};

export default Songs;
