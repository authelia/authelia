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

export const EFFECT_VERSION = "2025-11-10T01:00Z";

const HALO_COUNT = 6;
const SPARK_COUNT = 70;

function createHalos(width, height) {
  return Array.from({ length: HALO_COUNT }, (_, index) => {
    const base = (index + 1) / (HALO_COUNT + 1);
    return {
      x: width * 0.5,
      y: height * (0.18 + base * 0.6),
      radius: Math.max(width, height) * (0.12 + base * 0.18),
      thickness: 12 + index * 6,
      speed: 0.4 + index * 0.12,
      phase: Math.random() * Math.PI * 2,
    };
  });
}

function createSparks(width, height) {
  return Array.from({ length: SPARK_COUNT }, () => ({
    x: width * 0.5 + (Math.random() - 0.5) * width * 0.22,
    y: height * (0.25 + Math.random() * 0.5),
    radius: Math.random() * 1.6 + 0.6,
    driftX: (Math.random() - 0.5) * 0.4,
    driftY: Math.random() * 0.4 + 0.1,
    alpha: 0.4 + Math.random() * 0.4,
  }));
}

export function mount({ container }) {
  if (!container?.appendChild) {
    return undefined;
  }

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
  canvas.style.mixBlendMode = "screen";
  canvas.style.opacity = "0.88";

  container.appendChild(canvas);

  const state = {
    width: 0,
    height: 0,
    dpr: getDevicePixelRatio(),
    raf: 0,
    halos: [],
    sparks: [],
  };

  const resize = () => {
    const rect = container.getBoundingClientRect();
    state.dpr = getDevicePixelRatio();
    state.width = Math.max(0, rect.width);
    state.height = Math.max(0, rect.height);

    if (state.width === 0 || state.height === 0) {
      canvas.width = 1;
      canvas.height = 1;
      canvas.style.width = "1px";
      canvas.style.height = "1px";
      state.halos = [];
      state.sparks = [];
      return;
    }

    canvas.width = Math.floor(state.width * state.dpr);
    canvas.height = Math.floor(state.height * state.dpr);
    canvas.style.width = `${state.width}px`;
    canvas.style.height = `${state.height}px`;
    state.halos = createHalos(state.width, state.height);
    state.sparks = createSparks(state.width, state.height);
  };

  const draw = (timestamp) => {
    const { width, height, dpr, halos, sparks } = state;
    if (width === 0 || height === 0) {
      state.raf = requestFrame(draw);
      return;
    }

    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, width, height);

    const time = timestamp * 0.0011;

    for (const [index, halo] of halos.entries()) {
      const pulse = Math.sin(time * halo.speed + halo.phase) * 0.08 + 0.12;
      const radius = halo.radius * (1 + pulse * 0.4);
      const gradient = ctx.createRadialGradient(halo.x, halo.y, radius * 0.35, halo.x, halo.y, radius);
      gradient.addColorStop(0, "rgba(150, 200, 255, 0.08)");
      gradient.addColorStop(0.6, `rgba(140, 190, 255, ${0.2 + index * 0.05})`);
      gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

      ctx.lineWidth = halo.thickness + pulse * 40;
      ctx.strokeStyle = `rgba(140, 190, 255, ${0.06 + index * 0.03})`;
      ctx.beginPath();
      ctx.arc(halo.x, halo.y, radius, 0, Math.PI * 2);
      ctx.stroke();

      ctx.fillStyle = gradient;
      ctx.fillRect(halo.x - radius, halo.y - radius, radius * 2, radius * 2);
    }

    ctx.globalCompositeOperation = "lighter";

    for (const spark of sparks) {
      spark.y -= spark.driftY;
      spark.x += spark.driftX;

      if (spark.y < height * 0.12) {
        spark.y = height * 0.72 + Math.random() * height * 0.1;
        spark.x = width * 0.5 + (Math.random() - 0.5) * width * 0.18;
      }

      if (spark.x > width * 0.62) {
        spark.x = width * 0.38;
      } else if (spark.x < width * 0.38) {
        spark.x = width * 0.62;
      }

      const glow = ctx.createRadialGradient(spark.x, spark.y, 0, spark.x, spark.y, spark.radius * 4);
      glow.addColorStop(0, `rgba(170, 215, 255, ${spark.alpha})`);
      glow.addColorStop(0.5, `rgba(142, 185, 255, ${spark.alpha * 0.6})`);
      glow.addColorStop(1, "rgba(0, 0, 0, 0)");

      ctx.fillStyle = glow;
      ctx.beginPath();
      ctx.arc(spark.x, spark.y, spark.radius * 2.6, 0, Math.PI * 2);
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
