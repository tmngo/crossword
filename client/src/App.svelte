<script lang="ts">
  import { findClue } from './lib/crossword';

  import Grid from './lib/Grid.svelte';
  import type { Cell, Player, PlayerUpdate, Puzzle, PuzzleData, RawPuzzleData } from './lib/types';
  import { Direction, Tag, View } from './lib/types';

  let conn: WebSocket;
  let value = '';
  let view = View.LIST;
  let log = [];
  let savedPuzzles: PuzzleData[] = [];
  let puzzleMap = new Map<string, PuzzleData>(JSON.parse(localStorage.getItem('crosswords')));
  let puzzle: Puzzle = {
    acrossClues: [],
    attribution: '',
    creators: '',
    downClues: [],
    grid: 'Q',
    height: 1,
    id: '',
    title: '',
    width: 1,
  };
  let cells: Cell[] = [{ isCell: false, solution: 'Q', value: '' }];
  let playerMap: Map<string, Player>;
  let playerId = '';
  // $: playerIndex = players.findIndex((p: Player) => p.id === playerId);
  let activeClue = puzzle.acrossClues[0];
  console.log(view);
  let inputYear = 2021;
  let inputMonth = 7;
  let inputDay = 30;
  // $: {
  // if (playerIndex > 0) {
  // const { row, col, dir } = players[playerIndex].position;
  // activeClues = dir === Direction.ACROSS ? puzzle.acrossClues : puzzle.downClues;
  // activeClue = findClue(puzzle, players[playerIndex].position);
  // console.log({ activeClue });
  // }
  // }

  const onMessage = async (ev: MessageEvent<any>) => {
    console.log(ev);
    const text = await ev.data.text();
    const message = JSON.parse(text);
    const { tag, data } = message;

    switch (tag) {
      case Tag.Text:
        log = [...log, data];
        console.log(data);
        break;
      case Tag.Puzzle:
        puzzle = data as Puzzle;
        puzzle.width = puzzle.width;
        cells = Array(puzzle.grid.length);
        for (let i = 0; i < cells.length; i++) {
          const isCell = puzzle.grid[i] !== '.';
          cells[i] = {
            isCell,
            solution: puzzle.grid[i],
            value: isCell ? '' : '.',
          };
        }
        break;
      case Tag.State:
        console.log({ data });
        break;
      case Tag.Register:
        playerId = data.id;
        console.log({ playerId });
        break;
      case Tag.NEW_PUZZLE:
        const puzzles = data as RawPuzzleData[];
        for (const p of puzzles) {
          if (!puzzleMap.has(p.id)) {
            puzzleMap.set(p.id, {
              id: p.id,
              source: p.source,
              date: new Date(p.year, p.month - 1, p.day),
              title: p.title,
              state: p.state,
              completion: p.completion,
            } as PuzzleData);
          }
        }
        console.log({ puzzleMap });
        puzzleMap = puzzleMap;
        localStorage.setItem('crosswords', JSON.stringify([...puzzleMap]));
        break;
      case Tag.PLAYER_UPDATE:
        const { state, players: playerObj } = data as PlayerUpdate;
        console.log(state);
        console.log(message);
        for (let i = 0; i < cells.length && i < state.length; i++) {
          const isZero = state[i].charCodeAt(0) === 0;
          cells[i].value = isZero ? '' : state[i];
        }
        playerMap = new Map(Object.entries(playerObj));
        console.log(playerMap);
        console.log(Array.from(playerMap.values()));
        console.log(playerId);
        if (playerId !== '') {
          console.log(playerMap?.get(playerId).position);
        }
        break;
      default:
        console.log(`Undefined message tag: ${tag}.`);
        console.log(data);
        break;
    }
  };
  const connect = () => {
    if (window['WebSocket']) {
      // conn = new WebSocket("ws://" + document.location.host + "/ws");
      conn = new WebSocket('ws://' + 'localhost:8080' + '/ws' + document.location.pathname);
      conn.onopen = (ev: Event) => {
        console.log('Socket opened.');
      };
      conn.onclose = (ev: CloseEvent) => {
        console.log(
          'Connection closed. An attempt to reconnect will be made in 1 second.',
          ev.reason
        );
        setTimeout(connect, 1000);
      };
      conn.onmessage = onMessage;
    } else {
      console.log('Your browser does not support WebSockets.');
    }
  };

  const send = (tag: Tag, data: any) => {
    if (!conn) return;
    conn.send(JSON.stringify({ tag, data }));
  };

  connect();

  // const url = 'https://tmngo-proxy.herokuapp.com/herbach.dnsalias.com/wsj/wsj210731.puz';
  // "http://cdn.games.arkadiumhosted.com/latimes/assets/DailyCrossword/la190331.xml";
  // "https://tmngo-proxy.herokuapp.com/https://herbach.dnsalias.com/uc/ucs190331.puz";
  // "https://tmngo-proxy.herokuapp.com/https://herbach.dnsalias.com/uc/ucs200329.puz";
</script>

<svelte:window
  on:beforeunload={() => {
    if (conn) conn.close();
  }}
/>

<main>
  <div>
    {#if view === View.CROSSWORD}
      <div id="header">
        <span>{activeClue?.number ?? 'number'}</span>
        <span>{activeClue?.text ?? 'text'}</span>
        <span>[{activeClue?.length ?? 'length'}]</span>
      </div>
      <div id="puzzle-container">
        <Grid
          {playerId}
          {playerMap}
          {puzzle}
          {cells}
          on:player-action={(e) => {
            console.log({ e });
            send(Tag.PLAYER_ACTION, e.detail);
            const player = playerMap?.get(playerId);
            if (!player) return;
            activeClue = findClue(puzzle, player.position);
            // console.log({ activeClue });
          }}
          on:player-click={(e) => {
            send(Tag.PLAYER_CLICK, e.detail);
            const player = playerMap?.get(playerId);
            if (!player) return;
            activeClue = findClue(puzzle, player.position);
          }}
        />
      </div>
      <div id="clues">
        <div class="clue-header">Across</div>
        {#each puzzle.acrossClues as clue, i}
          <div class="clue-number">{clue.number}.</div>
          <div class="clue-text">{clue.text}</div>
        {/each}
        <div class="clue-header">Down</div>
        {#each puzzle.downClues as clue, i}
          <div class="clue-number">{clue.number}.</div>
          <div class="clue-text">{clue.text}</div>
        {/each}
      </div>
    {:else}
      <div class="crossword-list">
        {#each [...puzzleMap] as [id, puzzle]}
          <div class="title">“{puzzle.title}”</div>
          <div class="date">
            {new Date(puzzle.date).toLocaleDateString(undefined, {
              weekday: 'short',
              year: 'numeric',
              month: 'numeric',
              day: 'numeric',
            })}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <div id="controls" on:keydown|stopPropagation={() => {}}>
    <div id="log" style="display: flex">
      {#each log as entry, i}
        <div>{i}: {entry}</div>
      {/each}
    </div>
    <button
      on:click={() => {
        send(Tag.Puzzle, {
          sources: ['wsj'],
          day: 30,
          month: 7,
          year: 2021,
        });
      }}>WSJ Puzzle</button
    >

    <button
      on:click={() => {
        send(Tag.NEW_PUZZLE, {
          sources: ['wsj'],
          day: 30,
          month: 7,
          year: 2021,
        });
      }}>Get</button
    >
    <button
      on:click={() => {
        send(Tag.NEW_PUZZLE, {
          sources: ['wsj'],
          day: inputDay,
          month: inputMonth,
          year: inputYear,
        });
      }}>Get #</button
    >

    <button
      on:click={() => {
        if (!puzzle.id) return;
        const state = localStorage.getItem(puzzle.id);
        if (!state) return;
        send(Tag.PUZZLE_LOAD, {
          id: puzzle.id,
          state: state,
        });
        view = View.CROSSWORD;
      }}>Load</button
    >

    <input
      id="date-year"
      placeholder="YYYY"
      type="number"
      inputmode="numeric"
      pattern="[0-9]*"
      bind:value={inputYear}
    />
    <input
      id="date-month"
      placeholder="MM"
      type="number"
      inputmode="numeric"
      pattern="[0-9]*"
      bind:value={inputMonth}
    />
    <input
      id="date-day"
      placeholder="DD"
      type="number"
      inputmode="numeric"
      pattern="[0-9]*"
      bind:value={inputDay}
    />

    <button
      on:click={() => {
        if (!conn || !value) return;
        send(Tag.Text, value);
        value = '';
      }}>Send</button
    >
    <input type="text" bind:value />
  </div>
</main>

<style>
  :root {
    font-family: 'Atkinson Hyperlegible', sans-serif;
  }

  :global(html) {
    box-sizing: border-box;
  }

  :global(*),
  :global(*::before),
  :global(*::after) {
    box-sizing: inherit;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    overflow: hidden;
  }

  :global(#app) {
    height: 100vh;
  }

  @media screen and (orientation: portrait) {
    /* Portrait styles */
    main {
      flex-direction: column;
    }
  }

  @media screen and (orientation: landscape) {
    main {
      flex-direction: row;
    }
  }

  main {
    text-align: center;
    display: flex;
    justify-content: center;
    align-items: center;
    /* padding: 1em; */
    margin: 0 auto;
    position: relative;
    height: 100%;
  }

  #puzzle-container {
    background: #000;
    height: 90vw;
    width: 90vw;
    max-height: 90vh;
    max-width: 90vh;
  }

  #controls {
    background: #fffd;
    padding: 0.5em;
    position: absolute;
    bottom: 0;
    right: 0;
    margin: 1em;
  }

  #header {
    display: flex;
  }

  #header > * {
    margin-right: 0.5em;
  }

  #log {
    display: flex;
    flex-direction: column;
    outline: 1px solid black;
    align-items: flex-start;
    padding: 0.5em 0.75em;
    font-family: 'IBM Plex Mono';
  }

  #clues {
    max-height: 88%;
    max-width: 100%;
    font-size: 0.75em;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    overflow: auto;
    margin-left: 1em;
    display: grid;
    grid: auto-flow / auto 1fr;
    gap: 0 0.5em;
  }

  .clue-header {
    grid-column: span 2;
    font-weight: bold;
    margin-top: 0.5em;
  }

  .clue-number {
    grid-column: 1;
  }

  .clue-text {
    grid-column: 2;
  }

  button {
    font-family: 'Atkinson Hyperlegible', sans-serif;
    border-radius: 0;
    border: solid 1px #aaa;
    background: none;
    padding: 0.125em 0.5em;
    cursor: pointer;
  }

  input {
    padding: 0.125em 0.5em;
    font-family: 'Atkinson Hyperlegible';
  }

  .crossword-list {
    font-size: 2em;
    display: grid;
    grid: auto-flow / auto auto;
    gap: 0.5em 4em;
  }

  .crossword-list > .title {
    justify-self: flex-start;
  }

  .crossword-list > .date {
    justify-self: flex-end;
  }

  #date-year {
    width: 6em;
  }

  #date-month {
    width: 5em;
  }

  #date-day {
    width: 5em;
  }
</style>
