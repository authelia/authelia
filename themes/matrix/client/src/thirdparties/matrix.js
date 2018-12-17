// Parameters
const fontSize = 12;
const spdMult = 0.5;
const fadeSpd = 0.03;
const headColor = '#FFFFFF';
const tailColor = '#00FF00';

canvas.width  = window.innerWidth;
canvas.height = window.innerHeight;
let ctx = canvas.getContext('2d');
let pos, spd, time, chars;

function init() {
  pos = []; spd = []; time = []; chars = [];
  ctx.font = fontSize + 'pt Consolas';
  for (let i = 0; i < canvas.width / fontSize; i++) {
    pos[i] = Math.random() * (canvas.height / fontSize);
    spd[i] = (Math.random() + 0.2) * spdMult;
    time[i] = 0; 
    chars[i] = ' ';
  }
}

function render() {
  requestAnimationFrame(render);
    
  ctx.fillStyle = tailColor;
  for (let i = 0; i < chars.length; ++i) { // Tails
    ctx.fillText(chars[i], i * fontSize + 1, pos[i] * fontSize); 
  }
  ctx.fillStyle = `rgba(0, 0, 0, ${fadeSpd})`;
  ctx.fillRect(0, 0, canvas.width, canvas.height); // Fading

  ctx.fillStyle = headColor;
  for (let x = 0; x < pos.length; ++x){ // Chars
    if (time[x] > 1) {
      let charCode = (Math.random() < 0.9) ? Math.random() * 93 + 33
                                           : Math.random() * 15 + 12688;
      chars[x] = String.fromCharCode(charCode);
      ctx.fillText(chars[x], x * fontSize + 1, pos[x] * fontSize + fontSize);
      pos[x]++;
      if (pos[x] * fontSize > canvas.height) pos[x] = 0;
      time[x] = 0;
    }
    time[x] += spd[x];
  }
}

window.onload = function() {
  window.onresize = () => {
    canvas.width  = window.innerWidth;
    canvas.height = window.innerHeight;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    init();
  };
  init();
  render();
};
