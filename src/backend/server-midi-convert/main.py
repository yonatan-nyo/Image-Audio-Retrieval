from basic_pitch.inference import predict_and_save
from basic_pitch import ICASSP_2022_MODEL_PATH
import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from mido import MidiFile
import json
import uuid
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
        midi_data_array = convert_midi_to_array(midi_file_path)
        json_file_path = save_midi_array_to_json(midi_data_array, midi_file_path, os.path.basename(file_path))

        
        # Return the full path of the generated MIDI file
        logger.info(f"MIDI file created at: {full_midi_path}")
        return {"full_path": full_midi_path, "json_file_path": json_file_path,}
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

    # Call the function for predicting and saving
    predict_and_save(
        audio_files,             # List of input audio file paths
        output_dir,              # Output directory where MIDI files will be saved
        model_or_model_path=model_path,  # Path to the default model
        save_midi=True,          # Save as MIDI files
        sonify_midi=False,       # Sonify MIDI output
        save_model_outputs=False,  # Save intermediate model outputs
        save_notes=False         # Save note data as JSON
    )

    # Create the MIDI output file path by appending '_basic_pitch.mid'
    output_filename = f"{os.path.splitext(os.path.basename(file_path))[0]}_basic_pitch.mid"
    return os.path.join(output_dir, output_filename)

# Helper Function: Convert MIDI to Array
def convert_midi_to_array(midi_file_path: str):
    """
    Parse a MIDI file and convert its content to an array of integers.
    Each integer represents a MIDI note that is on.
    """
    midi_data_array = []
    midi = MidiFile(midi_file_path)

    for track in midi.tracks:
        for msg in track:
            if msg.type == "note_on" and msg.velocity > 0:
                midi_data_array.append(msg.note)

    return midi_data_array

# Helper Function: Save MIDI Array to JSON
def save_midi_array_to_json(midi_data_array, midi_file_path, original_filename):
    """
    Save the MIDI data array as a JSON file. Use the original file name with a unique identifier for the JSON file.
    """
    output_dir = os.path.dirname(midi_file_path)
    unique_id = uuid.uuid4().hex  # Generate a unique identifier
    json_filename = f"{os.path.splitext(original_filename)[0]}_{unique_id}_data.json"
    json_file_path = os.path.join(output_dir, json_filename)

    with open(json_file_path, "w") as json_file:
        json.dump(midi_data_array, json_file, indent=4)

    return json_file_path
