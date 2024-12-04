import { IAlbum } from "./Album";

export interface ISong {
  ID: number;
  Name: string;
  AudioFilePath: string;
  AudioFilePathMidi: string;

  AlbumID: number;
  Album: IAlbum;
}
