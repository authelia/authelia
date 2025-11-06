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

function lerp(a, b, t) {
  return a + (b - a) * t;
}

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
  canvas.style.opacity = "0.75";

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

  const drawWave = (time, amplitude, wavelength, speed, colorStops) => {
    const { width, height, pointerX } = state;
    ctx.beginPath();
    const baseY = lerp(height * 0.35, height * 0.6, state.pointerY);
    const offset = (time * speed + pointerX * 4) % TAU;
    const step = width / 120;

    ctx.moveTo(0, baseY);
    for (let x = 0; x <= width; x += step) {
      const angle = (x / wavelength) * TAU + offset;
      const y = baseY + Math.sin(angle) * amplitude;
      ctx.lineTo(x, y);
    }
    ctx.lineTo(width, height + 20);
    ctx.lineTo(0, height + 20);
    ctx.closePath();

    const gradient = ctx.createLinearGradient(0, baseY - amplitude, 0, baseY + amplitude * 2);
    colorStops.forEach(([stop, color]) => {
      gradient.addColorStop(stop, color);
    });
    ctx.fillStyle = gradient;
    ctx.globalCompositeOperation = "lighter";
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

    const t = timestamp * 0.00035;

    drawWave(t, height * 0.22, width * 0.9, 0.8, [
      [0, "rgba(255, 180, 120, 0.08)"],
      [0.6, "rgba(255, 144, 120, 0.24)"],
      [1, "rgba(255, 110, 90, 0.0)"],
    ]);

    drawWave(t * 1.2, height * 0.16, width * 0.7, 1.1, [
      [0, "rgba(122, 198, 255, 0.12)"],
      [0.5, "rgba(90, 174, 255, 0.16)"],
      [1, "rgba(60, 130, 255, 0.0)"],
    ]);

    drawWave(t * 0.6, height * 0.28, width * 1.2, 0.4, [
      [0, "rgba(255, 210, 132, 0.1)"],
      [0.4, "rgba(255, 190, 122, 0.24)"],
      [1, "rgba(255, 170, 110, 0.0)"],
    ]);

    ctx.restore();
    state.raf = requestFrame(draw);
  };

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
