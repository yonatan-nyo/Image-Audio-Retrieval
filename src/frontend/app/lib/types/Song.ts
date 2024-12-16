import { IAlbum } from "./Album";

export interface ISong {
  ID: number;
  Name: string;
  AudioFilePath: string;
  AudioFilePathMidi: string;
  SimilarityScore?: number;

  AlbumID: number;
  Album: IAlbum;
}
