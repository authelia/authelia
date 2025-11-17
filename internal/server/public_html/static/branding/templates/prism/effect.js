const globalScope = (() => {
  try {
    return globalThis;
  } catch {
    return undefined;
  }
})();

const documentInstance = globalScope?.document ?? null;

const requestFrame = globalScope?.requestAnimationFrame?.bind(globalScope);
const cancelFrame = globalScope?.cancelAnimationFrame?.bind(globalScope);
const addGlobalListener = globalScope?.addEventListener?.bind(globalScope);
const removeGlobalListener = globalScope?.removeEventListener?.bind(globalScope);

const getDevicePixelRatio = () => {
  const ratio = globalScope?.devicePixelRatio;
  return Number.isFinite(ratio) ? ratio : 1;
};

export const EFFECT_VERSION = "2025-11-03T01:30Z";

export function mount({ container }) {
  if (!documentInstance?.createElement || !requestFrame || !addGlobalListener || !removeGlobalListener) {
    return undefined;
  }

  const canvas = documentInstance.createElement("canvas");
  const ctx = canvas.getContext("2d");

  if (!ctx) {
    return undefined;
  }

  canvas.style.position = "absolute";
  canvas.style.inset = "0";
  canvas.style.pointerEvents = "none";
  canvas.style.opacity = "0.8";

  container.appendChild(canvas);

  const state = {
    width: 0,
    height: 0,
    dpr: getDevicePixelRatio(),
    raf: 0,
    pointerX: 0.5,
    pointerY: 0.5,
  };

  const resize = () => {
    const rect = container.getBoundingClientRect();
    state.dpr = getDevicePixelRatio();
    state.width = rect.width;
    state.height = rect.height;
    canvas.width = Math.max(1, Math.floor(rect.width * state.dpr));
    canvas.height = Math.max(1, Math.floor(rect.height * state.dpr));
    canvas.style.width = `${rect.width}px`;
    canvas.style.height = `${rect.height}px`;
  };

  const drawPrism = (time) => {
    const { width, height } = state;

    const cx = width * 0.5;
    const cy = height * 0.48;
    const beamCount = 8;
    const radius = Math.min(width, height) * 0.85;

    for (let i = 0; i < beamCount; i += 1) {
      const angle = (i / beamCount) * Math.PI + time * 0.6;

      const spread = 0.22 + Math.sin(time * 1.2 + i) * 0.05;
      const length = radius * (0.6 + Math.cos(time * 0.8 + i) * 0.12);

      const x1 = cx + Math.cos(angle - spread) * length;
      const y1 = cy + Math.sin(angle - spread) * length;
      const x2 = cx + Math.cos(angle + spread) * (length * 0.85);
      const y2 = cy + Math.sin(angle + spread) * (length * 0.85);

      const gradient = ctx.createLinearGradient(cx, cy, (x1 + x2) / 2, (y1 + y2) / 2);
      gradient.addColorStop(0, "rgba(145, 198, 255, 0.72)");
      gradient.addColorStop(0.45, "rgba(124, 186, 255, 0.45)");
      gradient.addColorStop(1, "rgba(64, 122, 255, 0)");

      ctx.beginPath();
      ctx.moveTo(cx, cy);
      ctx.lineTo(x1, y1);
      ctx.lineTo(x2, y2);
      ctx.closePath();
      ctx.fillStyle = gradient;
      ctx.globalAlpha = 0.65;
      ctx.fill();
    }

    const coreGradient = ctx.createRadialGradient(cx, cy, 0, cx, cy, radius * 0.35);
    coreGradient.addColorStop(0, "rgba(142, 206, 255, 0.8)");
    coreGradient.addColorStop(1, "rgba(38, 58, 94, 0)");

    ctx.beginPath();
    ctx.fillStyle = coreGradient;
    ctx.globalAlpha = 0.9;
    ctx.arc(cx, cy, radius * 0.35, 0, Math.PI * 2);
    ctx.fill();
  };

  const draw = (timestamp) => {
    const { width, height, dpr } = state;
    if (width === 0 || height === 0) {
      state.raf = requestFrame(draw);
      return;
    }

    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, width, height);

    ctx.globalCompositeOperation = "lighter";
    drawPrism(timestamp * 0.0012);

    ctx.restore();
    state.raf = requestFrame(draw);
  };

  resize();
  state.raf = requestFrame(draw);

  addGlobalListener("resize", resize, { passive: true });

  return () => {
    cancelFrame?.(state.raf);
    removeGlobalListener?.("resize", resize);
    canvas.remove();
  };
}

export default mount;
