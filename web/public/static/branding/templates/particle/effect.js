const globalScope = typeof globalThis === "undefined" ? undefined : globalThis;

const documentInstance =
  typeof globalScope !== "undefined" && "document" in globalScope ? globalScope.document : null;

const requestFrame =
  typeof globalScope !== "undefined" && typeof globalScope.requestAnimationFrame === "function"
    ? globalScope.requestAnimationFrame.bind(globalScope)
    : undefined;

const cancelFrame =
  typeof globalScope !== "undefined" && typeof globalScope.cancelAnimationFrame === "function"
    ? globalScope.cancelAnimationFrame.bind(globalScope)
    : undefined;

const addGlobalListener =
  typeof globalScope !== "undefined" && typeof globalScope.addEventListener === "function"
    ? globalScope.addEventListener.bind(globalScope)
    : undefined;

const removeGlobalListener =
  typeof globalScope !== "undefined" && typeof globalScope.removeEventListener === "function"
    ? globalScope.removeEventListener.bind(globalScope)
    : undefined;

const getDevicePixelRatio = () =>
  typeof globalScope !== "undefined" && typeof globalScope.devicePixelRatio === "number"
    ? globalScope.devicePixelRatio
    : 1;

const TAU = Math.PI * 2;
const PARTICLE_COUNT = 140;

function random(min, max) {
  return Math.random() * (max - min) + min;
}

const project = (x, y, z, width, height, pointerX, pointerY) => {
  const perspective = 1 / (1 - z * 0.75);
  const px = width * 0.5 + (x + (pointerX - 0.5) * 1.6) * width * 0.48 * perspective;
  const py = height * 0.45 + (y + (pointerY - 0.5) * 1.7) * height * 0.52 * perspective;
  const scale = perspective;
  return { x: px, y: py, scale };
};

export function mount({ container }) {
  if (
    !documentInstance ||
    typeof documentInstance.createElement !== "function" ||
    typeof requestFrame !== "function" ||
    typeof addGlobalListener !== "function" ||
    typeof removeGlobalListener !== "function"
  ) {
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
  canvas.style.opacity = "0.92";

  container.appendChild(canvas);

  const state = {
    width: 0,
    height: 0,
    dpr: getDevicePixelRatio(),
    raf: 0,
    pointerX: 0.5,
    pointerY: 0.5,
    smoothX: 0.5,
    smoothY: 0.5,
    particles: [],
  };

  const createParticles = () => {
    state.particles = Array.from({ length: PARTICLE_COUNT }, () => ({
      radius: random(0.35, 1.8),
      theta: random(0, TAU),
      phi: random(0, TAU),
      distance: random(0.2, 0.95),
      baseSpeed: random(0.0012, 0.0038),
      hue: random(185, 210),
      saturation: random(70, 88),
      lightness: random(68, 92),
      baseAlpha: random(0.22, 0.55),
      twinkle: random(0.6, 1.6),
    }));
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

  const draw = (timestamp) => {
    const { width, height, dpr, pointerX, pointerY, particles } = state;
    if (width === 0 || height === 0) {
      state.raf = requestFrame(draw);
      return;
    }

    ctx.save();
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, width, height);

    state.smoothX += (pointerX - state.smoothX) * 0.05;
    state.smoothY += (pointerY - state.smoothY) * 0.05;

    const time = timestamp * 0.0011;

    ctx.globalCompositeOperation = "lighter";

    particles.forEach((particle, index) => {
      const speed = particle.baseSpeed + (index / PARTICLE_COUNT) * 0.0009;
      const theta = particle.theta + time * speed * 90;
      const phi = particle.phi + time * speed * 65;

      const x = Math.cos(theta) * Math.sin(phi) * particle.distance;
      const y = Math.sin(theta) * Math.sin(phi) * particle.distance;
      const z = Math.cos(phi) * particle.distance;

      const { x: screenX, y: screenY, scale } = project(x, y, z, width, height, state.smoothX, state.smoothY);
      const radius = Math.max(0.5, particle.radius * scale * 1.6);
      const twinkle = Math.sin(time * particle.twinkle + index) * 0.45 + 0.75;

      const gradient = ctx.createRadialGradient(screenX, screenY, 0, screenX, screenY, radius * 3.4);
      const alpha = Math.min(1, particle.baseAlpha * (0.6 + twinkle));
      gradient.addColorStop(0, `hsla(${particle.hue}, ${particle.saturation}%, ${particle.lightness}%, ${alpha})`);
      gradient.addColorStop(
        0.4,
        `hsla(${particle.hue + 12}, ${particle.saturation - 8}%, ${particle.lightness + 6}%, ${alpha * 0.6})`,
      );
      gradient.addColorStop(1, "rgba(0, 0, 0, 0)");

      ctx.fillStyle = gradient;
      ctx.beginPath();
      ctx.arc(screenX, screenY, radius * 3.4, 0, TAU);
      ctx.fill();
    });

    ctx.restore();
    state.raf = requestFrame(draw);
  };

  createParticles();
  resize();
  state.raf = requestFrame(draw);

  addGlobalListener("resize", resize, { passive: true });

  return () => {
    if (typeof cancelFrame === "function") {
      cancelFrame(state.raf);
    }
    removeGlobalListener("resize", resize);
    if (canvas.parentElement === container) {
      container.removeChild(canvas);
    }
  };
}

export default mount;
