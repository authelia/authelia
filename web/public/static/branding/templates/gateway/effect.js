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

export const EFFECT_VERSION = "2025-11-13T02:05Z";

const POINTER_RADIUS = 150;
const LINK_RADIUS = 110;

function random(min, max) {
  return Math.random() * (max - min) + min;
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
  canvas.style.opacity = "0.85";
  canvas.style.mixBlendMode = "screen";

  container.appendChild(canvas);

  let width = 0;
  let height = 0;
  let dpr = getDevicePixelRatio();
  const particles = [];
  const pointer = { x: 0, y: 0, active: false };
  let animationFrame = 0;

  const resize = () => {
    const rect = container.getBoundingClientRect();
    width = rect.width;
    height = rect.height;
    dpr = getDevicePixelRatio();

    canvas.width = Math.max(1, Math.floor(width * dpr));
    canvas.height = Math.max(1, Math.floor(height * dpr));
    canvas.style.width = `${width}px`;
    canvas.style.height = `${height}px`;

    const targetCount = Math.max(140, Math.min(220, Math.floor((width * height) / 22000)));
    while (particles.length < targetCount) {
      particles.push({
        x: random(0, width),
        y: random(0, height),
        vx: random(-0.22, 0.22),
        vy: random(-0.22, 0.22),
        size: random(0.6, 1.4),
      });
    }
    if (particles.length > targetCount) {
      particles.length = targetCount;
    }
  };

  const updateParticle = (particle) => {
    particle.x += particle.vx;
    particle.y += particle.vy;

    if (particle.x <= 0 || particle.x >= width) {
      particle.vx *= -1;
      particle.x = Math.max(0, Math.min(width, particle.x));
    }
    if (particle.y <= 0 || particle.y >= height) {
      particle.vy *= -1;
      particle.y = Math.max(0, Math.min(height, particle.y));
    }

    particle.vx += random(-0.005, 0.005);
    particle.vy += random(-0.005, 0.005);

    if (pointer.active) {
      const dx = particle.x - pointer.x;
      const dy = particle.y - pointer.y;
      const distSq = dx * dx + dy * dy;
      if (distSq < POINTER_RADIUS * POINTER_RADIUS && distSq > 1) {
        const dist = Math.sqrt(distSq);
        const falloff = 1 - distSq / (POINTER_RADIUS * POINTER_RADIUS);
        const strength = falloff * falloff * 0.18;
        const nx = dx / dist;
        const ny = dy / dist;
        particle.vx += nx * strength;
        particle.vy += ny * strength;
      }
    }

    const speed = Math.hypot(particle.vx, particle.vy);
    const maxSpeed = 0.28;
    if (speed > maxSpeed) {
      const scale = maxSpeed / speed;
      particle.vx *= scale;
      particle.vy *= scale;
    }
  };

  const draw = () => {
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, width, height);

    ctx.globalAlpha = 0.85;
    ctx.fillStyle = "rgba(9, 8, 25, 0.62)";
    ctx.fillRect(0, 0, width, height);

    ctx.globalAlpha = 1;
    ctx.fillStyle = "rgba(120, 160, 255, 0.55)";
    for (const particle of particles) {
      updateParticle(particle);
      ctx.beginPath();
      ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2);
      ctx.fill();
    }

    const linkRadiusSq = LINK_RADIUS * LINK_RADIUS;
    ctx.lineWidth = 0.6;
    for (let i = 0; i < particles.length; i += 1) {
      const p = particles[i];
      for (let j = i + 1; j < particles.length; j += 1) {
        const q = particles[j];
        const dx = p.x - q.x;
        if (Math.abs(dx) > LINK_RADIUS) continue;
        const dy = p.y - q.y;
        if (Math.abs(dy) > LINK_RADIUS) continue;
        const distSq = dx * dx + dy * dy;
        if (distSq < linkRadiusSq) {
          const alpha = 1 - distSq / linkRadiusSq;
          ctx.globalAlpha = alpha * 0.6;
          ctx.beginPath();
          ctx.moveTo(p.x, p.y);
          ctx.lineTo(q.x, q.y);
          ctx.strokeStyle = "rgba(120, 160, 255, 0.45)";
          ctx.stroke();
        }
      }
    }
    ctx.globalAlpha = 1;

    animationFrame = requestFrame(draw);
  };

  const handlePointerMove = (event) => {
    const rect = container.getBoundingClientRect();
    pointer.x = event.clientX - rect.left;
    pointer.y = event.clientY - rect.top;
    pointer.active = true;
  };

  const handlePointerLeave = () => {
    pointer.active = false;
  };

  resize();
  animationFrame = requestFrame(draw);

  container.addEventListener("pointermove", handlePointerMove, { passive: true });
  container.addEventListener("pointerleave", handlePointerLeave, { passive: true });
  addGlobalListener("resize", resize, { passive: true });

  return () => {
    if (typeof cancelFrame === "function") {
      cancelFrame(animationFrame);
    }
    container.removeEventListener("pointermove", handlePointerMove);
    container.removeEventListener("pointerleave", handlePointerLeave);
    removeGlobalListener("resize", resize);
    if (canvas.parentElement === container) {
      container.removeChild(canvas);
    }
  };
}

export default mount;
