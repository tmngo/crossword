import { Clue, Direction, Position, Puzzle } from './types';

export const findClue = (puzzle: Puzzle, position: Position) => {
  const { row, col, dir } = position;
  const { acrossClues, downClues } = puzzle;
  if (dir === Direction.ACROSS) {
    return acrossClues.find((clue: Clue) => {
      return row === clue.row && col >= clue.column && col <= clue.column + clue.length;
    });
  }
  return downClues.find((clue: Clue) => {
    return col === clue.column && row >= clue.row && row <= clue.row + clue.length;
  });
};
