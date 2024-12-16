import { ISong } from "./Song";

export interface IAlbum {
  ID: number;
  Name: string;
  PicFilePath: string;
  Songs: ISong[];
}
