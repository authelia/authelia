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

export const EFFECT_VERSION = "2025-11-13T01:20Z";

const ARC_COUNT = 5;
const GLOW_COUNT = 18;

function createGlows(width, height) {
  return Array.from({ length: GLOW_COUNT }, () => ({
    x: width * (0.08 + Math.random() * 0.4),
    y: height * (0.15 + Math.random() * 0.65),
    radius: width * (0.08 + Math.random() * 0.12),
    hue: 240 + Math.random() * 40,
    saturation: 60 + Math.random() * 25,
    lightness: 60 + Math.random() * 20,
    alpha: 0.12 + Math.random() * 0.08,
    driftX: (Math.random() - 0.5) * 0.15,
    driftY: (Math.random() - 0.5) * 0.12,
  }));
}

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
  canvas.style.opacity = "0.95";
  canvas.style.mixBlendMode = "screen";

  container.appendChild(canvas);

  const state = {
    width: 0,
    height: 0,
    dpr: getDevicePixelRatio(),
    raf: 0,
    glows: [],
  };

  const resize = () => {
    const rect = container.getBoundingClientRect();
    state.dpr = getDevicePixelRatio();
    state.width = Math.max(1, rect.width);
    state.height = Math.max(1, rect.height);
    canvas.width = Math.floor(state.width * state.dpr);
    canvas.height = Math.floor(state.height * state.dpr);
    canvas.style.width = `${state.width}px`;
    canvas.style.height = `${state.height}px`;
    state.glows = createGlows(state.width, state.height);
  };

  const drawPanel = (width, height) => {
    const panelWidth = width * 0.62;
    const gradient = ctx.createLinearGradient(0, 0, panelWidth, height * 0.9);
    gradient.addColorStop(0, "rgba(132, 82, 255, 0.85)");
    gradient.addColorStop(0.4, "rgba(222, 94, 204, 0.75)");
    gradient.addColorStop(0.85, "rgba(255, 162, 162, 0.55)");
    gradient.addColorStop(1, "rgba(255, 186, 176, 0.3)");

    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.moveTo(-width * 0.1, 0);
    ctx.lineTo(panelWidth, 0);
    ctx.lineTo(panelWidth - width * 0.12, height);
    ctx.lineTo(-width * 0.2, height);
    ctx.closePath();
    ctx.fill();

    for (let i = 0; i < ARC_COUNT; i += 1) {
      const t = i / ARC_COUNT;
      const arcGradient = ctx.createLinearGradient(0, 0, panelWidth, height);
      arcGradient.addColorStop(0, `rgba(255, 255, 255, ${0.08 - t * 0.02})`);
      arcGradient.addColorStop(1, "rgba(255, 255, 255, 0)");

      ctx.strokeStyle = arcGradient;
      ctx.lineWidth = 2 + i * 2;
      ctx.beginPath();
      ctx.moveTo(-width * 0.1, height * (0.1 + t * 0.6));
      ctx.quadraticCurveTo(
        panelWidth * (0.6 + t * 0.25),
        height * (0.05 + t * 0.4),
        panelWidth - width * 0.12,
        height * (0.6 + t * 0.35),
      );
      ctx.stroke();
    }
  };

  const draw = (timestamp) => {
    const { width, height, dpr, glows } = state;
    if (width === 0 || height === 0) {
      state.raf = requestFrame(draw);
      return;
    }

    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, width, height);

    drawPanel(width, height);

    ctx.globalCompositeOperation = "lighter";
    const time = timestamp * 0.0013;

    for (const [index, glow] of glows.entries()) {
      glow.x += glow.driftX;
      glow.y += glow.driftY + Math.sin(time * (0.6 + index * 0.03)) * 0.2;

      const leftBound = -width * 0.05;
      const rightBound = width * 0.58;
      const topBound = height * 0.1;
      const bottomBound = height * 0.9;

      if (glow.x < leftBound || glow.x > rightBound) {
        glow.driftX *= -1;
      }
      if (glow.y < topBound || glow.y > bottomBound) {
        glow.driftY *= -1;
      }

      const gradient = ctx.createRadialGradient(glow.x, glow.y, 0, glow.x, glow.y, glow.radius);
      gradient.addColorStop(0, `hsla(${glow.hue}, ${glow.saturation}%, ${glow.lightness}%, ${glow.alpha})`);
      gradient.addColorStop(
        0.45,
        `hsla(${glow.hue + 30}, ${glow.saturation - 10}%, ${glow.lightness + 10}%, ${glow.alpha * 0.65})`,
      );
      gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

      ctx.fillStyle = gradient;
      ctx.beginPath();
      ctx.arc(glow.x, glow.y, glow.radius, 0, Math.PI * 2);
      ctx.fill();
    }

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
