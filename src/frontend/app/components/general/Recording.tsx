import React, { useState, useRef } from "react";
import Button from "./Button";

interface RecordingProps {
  onStop: () => void;
  searchByHumming: (audioBlob: Blob, filename: string) => Promise<void>;
}

const Recording: React.FC<RecordingProps> = ({ onStop, searchByHumming }) => {
  const [recording, setRecording] = useState(false);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioStreamRef = useRef<MediaStream | null>(null);

  // Function to start recording
  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      audioStreamRef.current = stream;
      mediaRecorderRef.current = new MediaRecorder(stream);

      mediaRecorderRef.current.start(5000);

      mediaRecorderRef.current.ondataavailable = (e: BlobEvent) => {
        console.log("Data available:", e.data);
        searchByHumming(e.data, "audio.wav");
      };
    } catch (err) {
      console.error("Error accessing microphone:", err);
    }
  };

  // Function to stop recording
  const stopRecording = () => {
    if (mediaRecorderRef.current) {
      mediaRecorderRef.current.stop();
      audioStreamRef.current?.getTracks().forEach((track) => track.stop());
      setRecording(false);
      onStop();
    }
  };

  return (
    <div className="flex gap-1">
      <Button onClick={startRecording} disabled={recording}>
        Start Listening
      </Button>
      <Button onClick={stopRecording} disabled={!recording} className="bg-red-600 hover:bg-red-700">
        Stop Listening
      </Button>
      <p>{recording ? "Recording..." : "Not recording"}</p>
    </div>
  );
};

export default Recording;
