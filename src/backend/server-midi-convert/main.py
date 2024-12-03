from basic_pitch.inference import predict_and_save
from basic_pitch import ICASSP_2022_MODEL_PATH
import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import logging


# Provide the default model path
model_path = ICASSP_2022_MODEL_PATH

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
    # Specify the audio file and output directory
    audio_files = [file_path]
    output_dir = os.path.dirname(file_path)
    # Call the function
    predict_and_save(
        audio_files,             # List of input audio file paths
        output_dir,              # Output directory where MIDI files will be saved
        model_or_model_path=model_path,  # Path to the default model
        save_midi=True,          # Save as MIDI files
        sonify_midi=False,        # Sonify MIDI output
        save_model_outputs=False,  # Save intermediate model outputs
        save_notes=False          # Save note data as JSON
    )

    return os.path.join(output_dir, f"{file_path}_basic_pitch.mid")
