<script lang="ts">
  import { fragmentShaderSource, vertexShaderSource } from './shaders';
  import { createProgram, createShader, debounce, debounceLeading, debounceTimer } from './lib';
  import { resizeCanvasToDisplaySize } from './utils';
  import { createEventDispatcher, onMount } from 'svelte';

  // import { atlas, atlas64, atlas128, atlas256 } from './msdf';
  // import { atlas256 } from "../assets/msdfmin256-32";
  // import { atlasObject } from '../assets/msdf64-4';
  import { atlasObject } from '../assets/msdf-32-2';
  import { Direction, Key } from './types';
  import type { Color, Player, Position } from './types';
  import { findClue } from './crossword';
  // import { atlas128 } from './msdf';

  //   const textureFile = "src/assets/msdf64.png";
  // const textureFile = 'src/assets/msdf64-4.png';
  const textureFile = 'src/assets/msdf-32-2.png';
  // const textureFile = "src/assets/msdf256.png";
  // const textureFile = "src/assets/msdfmin256-32.png";
  const { atlas: a, glyphs } = atlasObject;

  export let cells = [];
  export let puzzle;
  export let playerId;
  export let playerMap: Map<string, Player>;

  const dispatch = createEventDispatcher();

  interface State {
    i: number;
    j: number;
  }

  let triangleCount = 1;
  let mounted = false;
  let drawDebounced;
  let state: State = {
    i: -1,
    j: -1,
  };
  let dispatchTimer = { timer: undefined };

  $: {
    console.log('puzzle changed');
    initializeColors();
    triangleCount = puzzle.width * puzzle.height * 6;
    if (drawDebounced) drawDebounced();
  }

  let colorState: Color[] = [];

  $: {
    // players = players;
    colorState = colorState;
    initializeColors();
    updateColors(playerMap);
  }

  let canvas: HTMLCanvasElement;
  let gl: WebGL2RenderingContext;
  let drawScene: (gl: WebGL2RenderingContext) => void;

  /*
  • – a
  |
  b

  2 – 3   5
  | /   / |
  1   4 – 6
  */
  const generateQuad = (
    positions: number[],
    height: number,
    width: number,
    a0: number,
    a1: number,
    b0: number,
    b1: number
  ) => {
    const padding = (2 * 0.02) / width;
    const x0 = (1 - padding) * ((2 * a0) / width - 1);
    const x1 = (1 - padding) * ((2 * a1) / width - 1);
    const y0 = (1 - padding) * (1 - (2 * b1) / height);
    const y1 = (1 - padding) * (1 - (2 * b0) / height);
    positions.push(x0, y0, x0, y1, x1, y1, x0, y0, x1, y1, x1, y0);
  };

  const generateGridTriangles = () => {
    const positions = [];
    addColorQuads(positions, puzzle);
    addLetterQuads(positions, puzzle);
    addNumberQuads(positions, puzzle);
    return positions;
  };

  const addColorQuads = (positions: number[], puzzle: any) => {
    const { grid, height, width } = puzzle;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        const padding = 0.02;
        const a0 = j + padding;
        const a1 = j + 1 - padding;
        const b0 = i + padding;
        const b1 = i + 1 - padding;
        generateQuad(positions, height, width, a0, a1, b0, b1);
      }
    }
  };

  const addLetterQuads = (positions: number[], puzzle: any) => {
    const { grid, height, width } = puzzle;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        const hasBlock = grid[index] === '.';
        if (hasBlock) {
          generateQuad(positions, height, width, j, j + 1, i, i + 1);
          continue;
        }
        const {
          atlasBounds: { left, bottom, right, top },
        } = glyphs[cells[index].value];
        const scale = 0.75;
        const letterHeight = ((top - bottom - 2) / a.size) * scale;
        const letterWidth = ((right - left - 2) / a.size) * scale;
        const offset = 0.3;
        // x0, y0 is the bottom left
        const a0 = j + 0.5 - letterWidth / 2;
        const a1 = j + 0.5 + letterWidth / 2;
        const b0 = i + 1 - (1 + offset) * letterHeight;
        const b1 = i + 1 - offset * letterHeight;
        generateQuad(positions, height, width, a0, a1, b0, b1);
      }
    }
  };

  const addNumberQuads = (positions: number[], puzzle: any) => {
    const { grid, height, width } = puzzle;
    let clueNumber = 1;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        const hasBlock = grid[index] === '.';
        const hasBlockLeft = j === 0 || grid[index - 1] === '.';
        const hasBlockUp = i === 0 || grid[index - width] === '.';
        if (hasBlock || (!hasBlockLeft && !hasBlockUp)) {
          continue;
        }
        const digits = clueNumber.toString();
        const padding = 0.05;
        let a0 = j + padding;
        let a1 = j + padding;
        for (let k = 0; k < digits.length; k++) {
          const {
            atlasBounds: { left, bottom, right, top },
          } = glyphs[digits[k]];
          const letterHeight = (top - bottom) / a.size / 4;
          const letterWidth = (right - left) / a.size / 4;
          a1 += letterWidth;
          const b0 = i + padding;
          const b1 = i + padding + letterHeight;
          generateQuad(positions, height, width, a0, a1, b0, b1);
          a0 += letterWidth;
        }
        clueNumber += 1;
      }
    }
  };

  const generateGridTextureCoordinates = () => {
    // const { height, width, grid } = puzzle;
    const coordinates = [];
    addColorCoordinates(coordinates, puzzle);
    addLetterCoordinates(coordinates, puzzle);
    addNumberCoordinates(coordinates, puzzle);
    return coordinates;
  };

  const addColorCoordinates = (coordinates: number[], puzzle: any) => {
    const { grid, height, width } = puzzle;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        // const x0 = 265 / a.width;
        // const y0 = (a.width - 250) / a.width;
        // const x1 = 285 / a.width;
        // const y1 = (a.width - 260) / a.width;
        const {
          atlasBounds: { left, bottom, right, top },
        } = glyphs[''];
        const x0 = left / a.width;
        const y0 = (a.width - bottom) / a.width;
        const x1 = right / a.width;
        const y1 = (a.width - top) / a.width;
        coordinates.push(x0, y0, x0, y1, x1, y1, x0, y0, x1, y1, x1, y0);
        continue;
      }
    }
  };

  const addLetterCoordinates = (coordinates: number[], puzzle: any) => {
    const { grid, height, width } = puzzle;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        const letter = grid[index] === '.' ? '.' : cells[index].value;
        const {
          atlasBounds: { left, bottom, right, top },
        } = glyphs[letter];
        const x0 = left / a.width;
        const y0 = (a.width - bottom) / a.width;
        const x1 = right / a.width;
        const y1 = (a.width - top) / a.width;
        coordinates.push(x0, y0, x0, y1, x1, y1, x0, y0, x1, y1, x1, y0);
      }
    }
  };

  const addNumberCoordinates = (coordinates: number[], puzzle: any) => {
    const { height, width, grid } = puzzle;
    let clueNumber = 1;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        const hasBlock = grid[index] === '.';
        const hasBlockLeft = j === 0 || grid[index - 1] === '.';
        const hasBlockUp = i === 0 || grid[index - width] === '.';
        if (hasBlock || (!hasBlockLeft && !hasBlockUp)) {
          continue;
        }
        const digits = clueNumber.toString();
        for (let k = 0; k < digits.length; k++) {
          const {
            atlasBounds: { left, bottom, right, top },
          } = glyphs[digits[k]];
          let w = a.width;
          const offset = 0;
          const x0 = left / w;
          const y0 = (w - bottom - offset) / w;
          const x1 = right / w;
          const y1 = (w - top + offset) / w;
          coordinates.push(x0, y0, x0, y1, x1, y1, x0, y0, x1, y1, x1, y0);
        }
        clueNumber += 1;
      }
    }
  };

  const generateVertexColors = () => {
    const colors = [];
    addColorColors(colors, puzzle);
    addLetterColors(colors, puzzle);
    addNumberColors(colors, puzzle);
    return colors;
  };

  const addColorColors = (colors: number[], puzzle: any) => {
    const { height, width, grid } = puzzle;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        const color = colorState[index];
        const { r, g, b, a } = color;
        for (let k = 0; k < 6; k++) {
          colors.push(r, g, b, a);
        }
      }
    }
  };
  const addLetterColors = (colors: number[], puzzle: any) => {
    const { height, width, grid } = puzzle;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        // const color = colorState[index];
        const color = { r: 0, g: 0, b: 0, a: 0 };
        const { r, g, b, a } = color;
        for (let k = 0; k < 6; k++) {
          colors.push(r, g, b, a);
        }
      }
    }
  };
  const addNumberColors = (colors: number[], puzzle: any) => {
    const { height, width, grid } = puzzle;
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        const color = [0.0, 0.0, 0.0, 0.0];
        for (let k = 0; k < 6; k++) {
          colors.push(...color);
        }
      }
    }
  };

  const initializeColors = () => {
    puzzle.width = puzzle.width;
    const { height, width } = puzzle;
    colorState = Array(height * width);
    for (let i = 0; i < height; i++) {
      for (let j = 0; j < width; j++) {
        const index = i * width + j;
        colorState[index] = { r: 1.0, g: 1.0, b: 1.0, a: 1.0 };
      }
    }
  };

  onMount(() => {
    console.log('--- Begin onMount.');
    gl = canvas.getContext('webgl2');

    // Create shaders by uploading the GLSL source and compiling.
    const vertShader = createShader(gl, gl.VERTEX_SHADER, vertexShaderSource);
    const fragShader = createShader(gl, gl.FRAGMENT_SHADER, fragmentShaderSource);
    // Link the two shaders into a program.
    const program = createProgram(gl, vertShader, fragShader);

    const attributes = [
      { name: 'a_position', size: 2, normalized: false },
      { name: 'a_texcoord', size: 2, normalized: true },
      { name: 'a_color', size: 4, normalized: false },
    ];

    // Look up where the vertex data needs to go.
    // const positionAttributeLocation = gl.getAttribLocation(program, 'a_position');
    // const texcoordAttributeLocation = gl.getAttribLocation(program, 'a_texcoord');
    // const colorAttributeLocation = gl.getAttribLocation(program, 'a_color');

    const foregroundColorLocation = gl.getUniformLocation(program, 'u_color');
    const translationLocation = gl.getUniformLocation(program, 'u_translation');
    const SIZE_OF_FLOAT = 4;
    // console.log({ triangleCount });

    // Create a vertex array object (attribute state), and set it as active.
    const vao = gl.createVertexArray();
    gl.bindVertexArray(vao);

    const buffers = [];
    for (let i = 0; i < attributes.length; i++) {
      const { name, size, normalized } = attributes[i];
      // Look up where the vertex data needs to go.
      const location = gl.getAttribLocation(program, name);
      const buffer = gl.createBuffer();
      // Bind the buffer to the ARRAY_BUFFER target.
      gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
      if (false) {
        gl.bufferData(gl.ARRAY_BUFFER, 6 * triangleCount * SIZE_OF_FLOAT, gl.DYNAMIC_DRAW);
      }
      // Turn on the attribute.
      gl.enableVertexAttribArray(location);
      // Tell the attribute how to get data out of the buffer.
      gl.vertexAttribPointer(location, size, gl.FLOAT, normalized, 0, 0);
      buffers.push(buffer);
    }

    /* Vertex positions */

    // Create a buffer for three 2D (clip space) points.
    // const positionBuffer = gl.createBuffer();
    // Bind the buffer to the ARRAY_BUFFER target.
    // This makes the positionBuffer the active buffer.
    // gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
    // const positions = generateGridTriangles();
    // gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(positions), gl.STATIC_DRAW);
    // Turn on the attribute.
    // gl.enableVertexAttribArray(positionAttributeLocation);
    // Tell the attribute how to get data out of positionBuffer.
    // const size = 2; // Two values per position
    // const type = gl.FLOAT; // Each value is a float
    // const normalized = false; // Values are already between 0 and 1
    // const stride = 8; // 2 values per position, each value is 4 bytes
    // const offset = 0;
    // gl.vertexAttribPointer(positionAttributeLocation, size, type, normalized, stride, offset);

    /* Texture coordinates */

    // const texcoordBuffer = gl.createBuffer();
    // gl.bindBuffer(gl.ARRAY_BUFFER, texcoordBuffer);
    // const textureCoordinates = generateGridTextureCoordinates();
    // gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(textureCoordinates), gl.STATIC_DRAW);
    // gl.enableVertexAttribArray(texcoordAttributeLocation);
    // gl.vertexAttribPointer(
    //   texcoordAttributeLocation,
    //   2,
    //   gl.FLOAT,
    //   true, // convert from 0-255 to 0.0-1.0
    //   0, // Attributes are tightly packed, not interleaved.
    //   0
    // );

    /* Vertex colors */

    // const colorBuffer = gl.createBuffer();
    // gl.bindBuffer(gl.ARRAY_BUFFER, colorBuffer);
    // const colors = generateVertexColors();
    // gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(colors), gl.STATIC_DRAW);
    // gl.enableVertexAttribArray(colorAttributeLocation);
    // gl.vertexAttribPointer(colorAttributeLocation, 4, gl.FLOAT, false, 0, 0);

    const texture = gl.createTexture();
    gl.activeTexture(gl.TEXTURE0 + 0);
    gl.bindTexture(gl.TEXTURE_2D, texture);
    gl.texImage2D(
      gl.TEXTURE_2D,
      0,
      gl.RGBA,
      1,
      1,
      0,
      gl.RGBA,
      gl.UNSIGNED_BYTE,
      new Uint8Array([0, 0, 255, 255])
    );

    const image = new Image();
    image.src = textureFile;
    image.addEventListener('load', () => {
      gl.bindTexture(gl.TEXTURE_2D, texture);
      gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, image);
      gl.generateMipmap(gl.TEXTURE_2D);
      // console.log('Image loaded.');
      drawScene(gl);
    });

    gl.useProgram(program);
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

    drawScene = (gl: WebGL2RenderingContext) => {
      resizeCanvasToDisplaySize(gl.canvas as HTMLCanvasElement, 2);
      gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
      gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

      gl.bindVertexArray(vao);

      // Set attributes.
      const bufferData = [
        generateGridTriangles(), // 6948
        generateGridTextureCoordinates(), // 6948
        generateVertexColors(), // 16200
      ];

      for (let i = 0; i < buffers.length; i++) {
        gl.bindBuffer(gl.ARRAY_BUFFER, buffers[i]);
        // console.log(bufferData[i].length);
        if (false) {
          gl.bufferSubData(gl.ARRAY_BUFFER, 0, new Float32Array(bufferData[i]));
        } else {
          gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(bufferData[i]), gl.STATIC_DRAW);
        }
      }

      // Set uniforms.
      const uniformData = [
        [0.0, 0.0, 0.0],
        [0, 0],
      ];

      gl.uniform3fv(foregroundColorLocation, uniformData[0]);
      gl.uniform2fv(translationLocation, uniformData[1]);

      // Draw.
      const mode = gl.TRIANGLES;
      const first = 0; // Starting index in the array of vector points
      const count = bufferData[0].length / 2; // Number of indices to be rendered
      gl.drawArrays(mode, first, count);
    };

    drawScene(gl);
    mounted = true;
    console.log('--- End onMount.');
  });

  const setRowColor = (i: number, j: number, color: Color) => {
    const { width, grid } = puzzle;
    const index = i * width + j;
    if (grid[index] === '.') return;
    colorState[index] = {
      r: color.r * color.r,
      g: color.g * color.g,
      b: color.b * color.b,
      a: color.a,
    };
    let k = j - 1;
    while (puzzle.grid[i * width + k] !== '.' && k >= 0) {
      colorState[i * width + k] = color;
      k--;
    }
    k = j + 1;
    while (puzzle.grid[i * width + k] !== '.' && k < width) {
      colorState[i * width + k] = color;
      k++;
    }
  };

  drawDebounced = debounce(() => {
    drawScene(gl);
    console.log('draw');
  }, 1000);

  const onClick = (e: MouseEvent) => {
    const { grid, height, width } = puzzle;
    const { clientX, clientY } = e;
    const { clientWidth, clientHeight, offsetLeft, offsetTop } = canvas;
    const row = Math.floor(((clientY - offsetTop) / clientHeight) * height);
    const col = Math.floor(((clientX - offsetLeft) / clientWidth) * width);
    const player = playerMap?.get(playerId);
    if (!player) return;
    // console.log('before', player.position);
    if (grid[row * width + col] === '.') return;
    if (!setPlayerPosition(row, col, player?.position.dir)) return;
    // console.log('after', playerMap?.get(playerId).position);
    // if (colorState[i * width + j].r === 0) {
    //   setRowColor(i, j, { r: 1.0, g: 1.0, b: 1.0, a: 1.0 });
    // } else {
    //   setRowColor(i, j, { r: 0.0, g: 0.75, b: 0.75, a: 1.0 });
    // }
    dispatch('player-click', { row, col });
    drawScene(gl);
  };

  const setPlayerPosition = (row: number, col: number, dir: Direction) => {
    const { grid, height, width } = puzzle;
    if (row < 0 || col < 0 || row >= height || col >= width) return false;

    const player = playerMap?.get(playerId);
    if (!player) return false;
    const {
      position: { row: currentRow, col: currentCol },
    } = player;

    console.log({ player });

    while (grid[row * width + col] === '.') {
      if (col === currentCol) {
        if (row > currentRow && row < height - 1) {
          row += 1;
        } else if (row < currentRow && row > 0) {
          row -= 1;
        } else {
          return false;
        }
      } else if (row === currentRow) {
        if (col > currentCol && col < width - 1) {
          col += 1;
        } else if (col < currentCol && col > 0) {
          col -= 1;
        } else {
          return false;
        }
      }
    }

    playerMap.set(playerId, {
      ...player,
      position: { row, col, dir },
    });
    updateColors(playerMap);
    playerMap = playerMap;
    console.log({ playerMap });
    return true;
  };

  const onKeyDown = (e: KeyboardEvent) => {
    console.log({ e, state });
    const { row, col, dir } = playerMap?.get(playerId).position;
    if (row === -1 || col === -1) return;
    const { width } = puzzle;
    const index = row * width + col;

    const { key } = e;
    switch (key) {
      case Key.SPACE:
        if (!setPlayerPosition(row, col, 1 - dir)) return;
        break;
      case Key.BACKSPACE:
      case Key.DELETE:
        cells[index].value = '';
        break;
      case Key.ARROW_DOWN:
        if (!setPlayerPosition(row + 1, col, dir)) return;
        break;
      case Key.ARROW_UP:
        if (!setPlayerPosition(row - 1, col, dir)) return;
        break;
      case Key.ARROW_LEFT:
        if (!setPlayerPosition(row, col - 1, dir)) return;
        break;
      case Key.ARROW_RIGHT:
        if (!setPlayerPosition(row, col + 1, dir)) return;
        break;
      default:
        if (key.length === 1) {
          const code = key.charCodeAt(0);
          console.log({ code });
          if (code < 97 || code > 122) return;
          cells[index].value = key.toUpperCase();
          if (dir === Direction.ACROSS) {
            setPlayerPosition(row, col + 1, dir);
          } else {
            setPlayerPosition(row + 1, col, dir);
          }
        }
    }

    dispatch('player-action', key);

    // dispatch('player-action', key);
    drawScene(gl);
  };

  const updateColors = (playerMap: Map<string, Player>) => {
    console.log('updateColors');
    if (!playerMap) return;
    const { width } = puzzle;
    for (const [_, p] of playerMap) {
      // console.log(p);
      const { color, position } = p;
      const { row, col, dir } = position;
      if (row < 0 || col < 0 || puzzle.acrossClues.length === 0) continue;
      const clue = findClue(puzzle, position);
      for (let i = 0; i < clue.length; i++) {
        const index =
          dir === Direction.ACROSS
            ? clue.row * width + clue.column + i
            : (clue.row + i) * width + clue.column;
        colorState[index] = color;
      }
      // colorState[row * width + col] = { r: 0.625, g: 0.875, b: 0.975, a: 1.0 };
      colorState[row * width + col] = {
        r: Math.pow(color.r, 0.3),
        g: Math.pow(color.g, 0.3),
        b: Math.pow(color.b, 0.3),
        a: 1.0,
      };
    }
    drawScene?.(gl);
  };
</script>

<svelte:window
  on:keydown={onKeyDown}
  on:resize={() => {
    drawDebounced();
  }}
/>

<canvas id="canvas" bind:this={canvas} on:click={onClick} />

<style>
  #canvas {
    height: 100%;
    width: 100%;
  }
</style>
