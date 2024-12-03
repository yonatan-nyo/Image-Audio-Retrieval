import librosa
import pretty_midi
import numpy as np
import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging

logger = logging.getLogger('uvicorn.error')

app = FastAPI()


class FilePathRequest(BaseModel):
    file_path: str


@app.post("/convert-to-midi/")
async def convert_to_midi(request: FilePathRequest):
    """
    Convert an audio file from the given path to MIDI format.
    """
    file_path = f"..\go\{request.file_path}"
    logger.info(f"Received file path: {file_path}")

    if not os.path.exists(file_path):
        logger.error(f"File not found: {file_path}")
        raise HTTPException(status_code=404, detail="File not found")

    try:
        # Generate MIDI file
        midi_file_path = convert_audio_to_midi(file_path)
        full_midi_path = os.path.abspath(midi_file_path)

        # Return the full path of the generated MIDI file
        logger.info(f"MIDI file created at: {full_midi_path}")
        return {"full_path": full_midi_path}
    except Exception as e:
        logger.error(f"Error processing file: {str(e)}")
        raise HTTPException(
            status_code=500, detail=f"Error processing file: {str(e)}"
        )


def convert_audio_to_midi(file_path: str) -> str:
    """
    Convert audio file to MIDI format using pitch detection and MIDI creation.
    """
    # Load the audio file using librosa
    try:
        audio, sr = librosa.load(file_path, sr=None)
    except Exception as e:
        raise HTTPException(
            status_code=400, detail=f"Error loading audio file: {str(e)}")

    try:
        # Use Harmonic Product Spectrum (HPS) or other pitch detection techniques
        # to detect pitches
        pitches, magnitudes = librosa.core.piptrack(y=audio, sr=sr)

        # Create a pretty_midi object
        midi = pretty_midi.PrettyMIDI()

        # Use program number directly for the instrument (e.g., 0 is Acoustic Grand Piano)
        instrument = pretty_midi.Instrument(
            program=0)  # 0 = Acoustic Grand Piano

        # Traverse the pitch detection results and create notes
        for t in range(pitches.shape[1]):
            # Find the pitch with the maximum magnitude
            index = magnitudes[:, t].argmax()
            pitch = pitches[index, t]

            if pitch > 0:  # Filter out non-pitch values
                # Convert pitch to MIDI note number (clamp it to valid range 0-127)
                midi_note = frequency_to_midi(pitch)
                if midi_note is None:
                    continue  # Skip invalid pitches

                time = librosa.frames_to_time(t, sr=sr)
                note = pretty_midi.Note(
                    velocity=100, pitch=midi_note, start=time, end=time + 0.1)
                instrument.notes.append(note)

        # Add the instrument to the MIDI object
        midi.instruments.append(instrument)

        # Save the MIDI file
        midi_file_path = file_path.rsplit(".", 1)[0] + ".mid"
        midi.write(midi_file_path)

        return midi_file_path
    except Exception as e:
        raise HTTPException(
            status_code=500, detail=f"Error processing audio file: {str(e)}")


def frequency_to_midi(frequency: float) -> int:
    """
    Convert a frequency (in Hz) to a MIDI note number.
    """
    if frequency <= 0:
        return None

    # Convert frequency to MIDI note number (clamp it between 0 and 127)
    midi_note = 69 + 12 * np.log2(frequency / 440.0)

    # Ensure the MIDI note is within the valid range (0 to 127)
    if midi_note < 0:
        return None
    elif midi_note > 127:
        return 127
    else:
        return int(round(midi_note))
