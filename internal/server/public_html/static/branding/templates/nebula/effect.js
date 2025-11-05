/* eslint-disable */
const TWO_PI = Math.PI * 2;

const PETAL_COUNT = 54;
const BASE_RADIUS = 0.32;

function random(min, max) {
  return Math.random() * (max - min) + min;
}

export function mount({ container }) {
  console.log("[NebulaEffect] mount v2025-11-03T01:45Z");
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");

  if (!ctx) {
    return undefined;
  }

  canvas.style.position = "absolute";
  canvas.style.inset = "0";
  canvas.style.pointerEvents = "none";
  canvas.style.opacity = "0.85";
  canvas.style.mixBlendMode = "screen";

  container.appendChild(canvas);

  const state = {
    width: 0,
    height: 0,
    dpr: window.devicePixelRatio || 1,
    raf: 0,
    pointerX: 0.5,
    pointerY: 0.5,
    smoothX: 0.5,
    smoothY: 0.5,
    petals: [],
  };

  const createPetals = () => {
    state.petals = Array.from({ length: PETAL_COUNT }, () => ({
      orbitRadius: random(0.22, 0.78),
      size: random(0.12, 0.26),
      thickness: random(0.08, 0.16),
      hue: random(208, 312),
      saturation: random(68, 90),
      lightness: random(60, 82),
      alpha: random(0.28, 0.6),
      angleOffset: random(0, TWO_PI),
      driftSpeed: random(0.12, 0.38),
      wobble: random(0.8, 1.6),
    }));
  };

  const resize = () => {
    const rect = container.getBoundingClientRect();
    state.dpr = window.devicePixelRatio || 1;
    state.width = rect.width;
    state.height = rect.height;
    canvas.width = Math.max(1, Math.floor(rect.width * state.dpr));
    canvas.height = Math.max(1, Math.floor(rect.height * state.dpr));
    canvas.style.width = `${rect.width}px`;
    canvas.style.height = `${rect.height}px`;
  };

  const drawPetal = (x, y, angle, size, thickness, hue, saturation, lightness, alpha) => {
    ctx.save();
    ctx.translate(x, y);
    ctx.rotate(angle);

    const gradient = ctx.createLinearGradient(0, 0, size * 60, 0);
    gradient.addColorStop(0, `hsla(${hue}, ${saturation}%, ${lightness}%, ${alpha})`);
    gradient.addColorStop(0.6, `hsla(${hue + 20}, ${saturation - 12}%, ${lightness + 8}%, ${alpha * 0.65})`);
    gradient.addColorStop(1, "rgba(255, 255, 255, 0)");

    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.quadraticCurveTo(size * 30, -thickness * 25, size * 60, 0);
    ctx.quadraticCurveTo(size * 30, thickness * 25, 0, 0);
    ctx.closePath();
    ctx.globalCompositeOperation = "lighter";
    ctx.fill();

    ctx.restore();
  };

  const draw = (timestamp) => {
    const { width, height, dpr, petals } = state;
    if (width === 0 || height === 0) {
      state.raf = requestAnimationFrame(draw);
      return;
    }

    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, width, height);

    // Smooth pointer easing (hover anchors still contribute subtly).
    state.smoothX += (state.pointerX - state.smoothX) * 0.05;
    state.smoothY += (state.pointerY - state.smoothY) * 0.05;

    const time = timestamp * 0.00035;
    const baseX = width * (0.5 + (state.smoothX - 0.5) * 0.25);
    const baseY = height * (0.48 + (state.smoothY - 0.5) * 0.3);

    ctx.globalCompositeOperation = "lighter";

    petals.forEach((petal, index) => {
      const orbitAngle = petal.angleOffset + time * petal.driftSpeed + index * 0.08;
      const wobble = Math.sin(time * 2 + index) * petal.wobble * 0.12;

      const px = baseX + Math.cos(orbitAngle) * width * (BASE_RADIUS + petal.orbitRadius * 0.6);
      const py = baseY + Math.sin(orbitAngle) * height * (BASE_RADIUS * 0.5 + petal.orbitRadius * 0.38);

      const angle = orbitAngle + Math.sin(time * 1.4 + index) * 0.6;
      const size = petal.size * (1 + Math.sin(time * 1.8 + index * 0.5) * 0.2);

      drawPetal(
        px,
        py,
        angle + wobble,
        size,
        petal.thickness,
        petal.hue,
        petal.saturation,
        petal.lightness,
        petal.alpha
      );
    });

    ctx.restore();
    state.raf = requestAnimationFrame(draw);
  };

  createPetals();
  resize();
  state.raf = requestAnimationFrame(draw);

  window.addEventListener("resize", resize);

  return () => {
    cancelAnimationFrame(state.raf);
    window.removeEventListener("resize", resize);
    if (canvas.parentElement === container) {
      container.removeChild(canvas);
    }
  };
}

export default mount;
