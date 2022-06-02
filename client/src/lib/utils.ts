/**
 * Resize a canvas to match the size its displayed.
 */
export const resizeCanvasToDisplaySize = (
  canvas: HTMLCanvasElement,
  multiplier?: number
): boolean => {
  multiplier = multiplier || 1;
  const width = (canvas.clientWidth * multiplier) | 0;
  const height = (canvas.clientHeight * multiplier) | 0;
  if (canvas.width !== width || canvas.height !== height) {
    canvas.width = width;
    canvas.height = height;
    return true;
  }
  return false;
}