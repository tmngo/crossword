export const enum Tag {
  Text,
  Puzzle,
  Register,
  State,
  PLAYER_UPDATE,
  PLAYER_ACTION,
  PLAYER_CLICK,
  PUZZLE_LOAD,
  NEW_PUZZLE,
}

export const enum Source {
  WALL_STREET_JOURNAL,
}

export interface Puzzle {
  acrossClues: Clue[];
  attribution: string;
  creators: string;
  downClues: Clue[];
  grid: string;
  height: number;
  id: string;
  title: string;
  width: number;
}

export interface RawPuzzleData {
  source: string;
  year: number;
  month: number;
  day: number;
  id: string;
  completion: number;
  title: string;
  state: string;
}

export interface PuzzleData {
  source: string;
  date: Date;
  id: string;
  completion: number;
  title: string;
  state: string;
}
export interface Clue {
  number: number;
  row: number;
  column: number;
  length: number;
  text: string;
}

export interface Cell {
  isCell: boolean;
  solution: string;
  value: string;
}

export interface Crossword {
  width: number;
  height: number;
  numClues: number;
  gridString: string;
}

export const enum Direction {
  ACROSS,
  DOWN,
}

export const enum Key {
  ARROW_DOWN = 'ArrowDown',
  ARROW_LEFT = 'ArrowLeft',
  ARROW_UP = 'ArrowUp',
  ARROW_RIGHT = 'ArrowRight',
  BACKSPACE = 'Backspace',
  DELETE = 'Delete',
  SPACE = ' ',
}

export interface Position {
  row: number;
  col: number;
  dir: Direction;
}

export interface Player {
  name: string;
  id: string;
  color: Color;
  position: Position;
}

export interface PlayerUpdate {
  state: string;
  players: { [index: string]: Player };
}

export interface Color {
  r: number;
  g: number;
  b: number;
  a: number;
}

export const enum View {
  CROSSWORD,
  LIST,
}
