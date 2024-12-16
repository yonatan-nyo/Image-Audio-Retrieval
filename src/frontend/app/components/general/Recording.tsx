import React, { useState, useRef } from "react";
import Button from "./Button";

interface RecordingProps {
  searchByHumming: (audioBlob: Blob, filename: string) => Promise<void>;
}

const Recording: React.FC<RecordingProps> = ({ searchByHumming }) => {
  const [recording, setRecording] = useState(false);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioStreamRef = useRef<MediaStream | null>(null);
  const recordingRef = useRef<boolean>(false); // Ref to store the current recording state

  // Function to start recording
  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      audioStreamRef.current = stream;
      mediaRecorderRef.current = new MediaRecorder(stream);

      mediaRecorderRef.current.ondataavailable = async (e: BlobEvent) => {
        try {
          await searchByHumming(e.data, "audio.wav");
        } catch (error) {
          console.error("not found or error");
        }
      };

      // Start recording
      mediaRecorderRef.current.start();
      setRecording(true);
      recordingRef.current = true;

      // Recursively stop and start recording every 5 seconds
      const stopAndRestartRecording = () => {
        if (mediaRecorderRef.current && mediaRecorderRef.current.state === "recording") {
          console.log("Stopping recording...");
          mediaRecorderRef.current.stop();

          // Wait for the stop event before restarting
          setTimeout(() => {
            if (recordingRef.current) {
              console.log("Restarting recording...");
              mediaRecorderRef.current?.start();
              setTimeout(stopAndRestartRecording, 5000); // Recursively restart every 5 seconds
            }
          }, 1000); // Wait a bit for the stop to complete before restarting
        }
      };

      // Start the recursive cycle
      setTimeout(stopAndRestartRecording, 5000);
    } catch (err) {
      console.error("Error accessing microphone:", err);
    }
  };

  return (
    <div className="flex gap-2 items-center">
      <Button
        onClick={() => {
          setRecording(true);
          recordingRef.current = true; // Update the ref as well
          startRecording();
        }}
        disabled={recording}>
        Start Listening
      </Button>
      {recording && (
        <Button
          onClick={() => {
            setRecording(false);
            recordingRef.current = false; // Update the ref when stopping
          }}
          disabled={!recording}
          className="bg-red-600 hover:bg-red-700">
          Stop Listening
        </Button>
      )}
      <p>{recording ? "Recording..." : "Not recording"}</p>
    </div>
  );
};

export default Recording;
