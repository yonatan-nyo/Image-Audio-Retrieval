import { useState, useRef, ChangeEvent } from "react";
import Button from "~/components/general/Button";
import axiosInstance from "~/utils/axiosInstance";

const AlbumsUpload: React.FC = () => {
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [uploadProgress, setUploadProgress] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const handleFileSelect = (event: ChangeEvent<HTMLInputElement>): void => {
    if (event.target.files) {
      setSelectedFiles([...selectedFiles, ...Array.from(event.target.files)]);
    }
  };

  const handleRemoveFile = (index: number): void => {
    setSelectedFiles(selectedFiles.filter((_, i) => i !== index));
  };

  const uploadFile = async (file: File): Promise<void> => {
    const formData = new FormData();
    formData.append("file", file);

    setLoading(true);
    setUploadProgress(`Uploading ${file.name}...`);

    try {
      const response = await axiosInstance.post("/albums/upload", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (response.status === 200) {
        setUploadProgress(`${file.name} uploaded successfully!`);
      } else {
        setUploadProgress(`Failed to upload ${file.name}`);
      }
    } catch (error) {
      console.error("Error uploading file:", error);
      setUploadProgress(`Error uploading ${file.name}`);
    } finally {
      setLoading(false);
    }
  };

  const handleUploadAll = async (): Promise<void> => {
    if (selectedFiles.length === 0) {
      alert("Please select files to upload.");
      return;
    }

    for (const file of selectedFiles) {
      await uploadFile(file);
    }

    setSelectedFiles([]); // Clear selected files after upload
    setUploadProgress(null); // Reset upload progress
  };

  return (
    <main className="p-4 w-0 grow">
      <h1 className="mb-4 text-xl font-semibold">Upload Album</h1>

      <div className="border p-4 rounded-lg shadow-md">
        <div className="mb-4">
          <Button onClick={() => fileInputRef.current?.click()}>Select Files</Button>
          <input
            type="file"
            accept=".zip,.png,.jpg,.jpeg,.webp" // Supporting images, audio, and zip files
            multiple
            ref={fileInputRef}
            onChange={handleFileSelect}
            className="hidden"
          />
        </div>

        {/* Display Selected Files */}
        {selectedFiles.length > 0 && (
          <div className="mb-4">
            <h2 className="text-lg font-medium mb-2">Selected Files:</h2>
            <ul className="list-disc list-inside">
              {selectedFiles.map((file, index) => (
                <li key={index} className="flex justify-between items-center bg-gray-100 text-black p-2 rounded mb-1">
                  <span>{file.name}</span>
                  <Button className="text-red-500" onClick={() => handleRemoveFile(index)}>
                    X
                  </Button>
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Upload Button */}
        <div className="mb-4">
          <Button onClick={handleUploadAll} disabled={selectedFiles.length === 0 || loading}>
            Upload
          </Button>
        </div>

        {/* Upload Progress */}
        {uploadProgress && <div className="mt-2 text-blue-500">{uploadProgress}</div>}
      </div>
    </main>
  );
};

export default AlbumsUpload;
