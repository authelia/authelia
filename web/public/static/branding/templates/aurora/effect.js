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
  const context = canvas.getContext("2d");

  if (!context) {
    return undefined;
  }

  canvas.style.width = "100%";
  canvas.style.height = "100%";
  canvas.style.pointerEvents = "none";
  canvas.style.position = "absolute";
  canvas.style.inset = "0";
  canvas.style.opacity = "0.85";

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

  const draw = (timestamp) => {
    const { width, height, dpr, pointerX, pointerY } = state;

    if (width === 0 || height === 0) {
      state.raf = requestFrame(draw);
      return;
    }

    context.save();
    context.setTransform(dpr, 0, 0, dpr, 0, 0);
    context.clearRect(0, 0, width, height);

    const time = timestamp * 0.0004;
    const waveX = Math.sin(time) * width * 0.18 + (pointerX - 0.5) * width * 0.3;
    const waveY = Math.cos(time * 0.8) * height * 0.22 + (pointerY - 0.5) * height * 0.35;

    const centerX = width * 0.5 + waveX;
    const centerY = height * 0.35 + waveY;

    const gradient = context.createRadialGradient(centerX, centerY, width * 0.05, centerX, centerY, width * 0.85);
    gradient.addColorStop(0, "rgba(146, 214, 255, 0.75)");
    gradient.addColorStop(0.35, "rgba(120, 152, 255, 0.55)");
    gradient.addColorStop(0.72, "rgba(30, 40, 96, 0.05)");
    gradient.addColorStop(1, "rgba(4, 10, 28, 0)");

    context.globalCompositeOperation = "lighter";
    context.fillStyle = gradient;
    context.fillRect(0, 0, width, height);

    const ribbonCount = 3;
    for (let i = 0; i < ribbonCount; i += 1) {
      const phase = time * (1 + i * 0.18) + i * Math.PI * 0.5;
      const offsetX = Math.sin(phase) * width * 0.25 + (pointerX - 0.5) * width * 0.2;
      const offsetY = Math.cos(phase * 0.9) * height * 0.12 + (pointerY - 0.5) * height * 0.18;
      const ribbonGradient = context.createLinearGradient(0, 0, width, height);
      ribbonGradient.addColorStop(0, "rgba(102, 210, 255, 0.18)");
      ribbonGradient.addColorStop(0.5, "rgba(168, 134, 255, 0.22)");
      ribbonGradient.addColorStop(1, "rgba(255, 176, 206, 0)");

      context.save();
      context.translate(width * 0.5, height * 0.5);
      context.rotate(Math.sin(phase) * 0.35);
      context.translate(-width * 0.5, -height * 0.5);
      context.globalAlpha = 0.45 - i * 0.1;
      context.fillStyle = ribbonGradient;
      context.fillRect(offsetX - width * 0.4, offsetY - height * 0.5, width * 1.4, height * 1.2);
      context.restore();
    }

    context.restore();
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
