const WIDTH = 40;
const HEIGHT = 25;

const C64Colors = [
    '#000000',
    '#FFFFFF',
    '#67372B',
    '#6FA3B1',

    '#6F3C85',
    '#588C43',
    '#342879',
    '#B7C66E',

    '#6F4F25',
    '#423900',
    '#996659',
    '#434343',

    '#6B6B6B',
    '#9AD183',
    '#6B5EB4',
    '#959595',
];

function main() {
    let canvas = {
        element: document.getElementById('canvas'),
        scalingFactor: 1,
        font: loadFont(),
        buffer: (function() {
            let a = new Array(HEIGHT);
            for (let i = 0; i < HEIGHT; i++) {
                a[i] = new Array(WIDTH);
                for (let j = 0; j < WIDTH; j++) {
                    a[i][j] = {char: 32, bg: 0, fg: 1, r: false}
                }
            }
            return a;
        })(),
        cursorPosition: {row: 0, col: 0},
        color: {fg: 1, bg: 0}
    };

    initializeColorSelectors(canvas);
    initializeCanvas(canvas);
    resizeCanvas(canvas);
    initializeWebSocket(canvas);
    setTimeout(function() {
        renderBuffer(canvas);
    }, 0);
}

function loadFont() {
    let font = new Image();
    font.src = '/static/font.png';
    return font
}

function initializeColorSelectors(canvas) {
    let header1 = document.getElementById('bg-colors');
    let header2 = document.getElementById('fg-colors');

    for (let i = 0; i < C64Colors.length; ++i) {
        let el1 = document.createElement('span');
        el1.classList.add('color-selector');
        el1.id = 'fg' + i;
        el1.style.backgroundColor = C64Colors[i];
        el1.addEventListener('click', function() {
            setForegroundColor(i, canvas);
        });
        header1.appendChild(el1);

        let el2 = document.createElement('span');
        el2.classList.add('color-selector');
        el2.id = 'bg' + i;
        el2.style.backgroundColor = C64Colors[i];
        el2.addEventListener('click', function() {
            setBackgroundColor(i, canvas);
        });
        header2.appendChild(el2);
    }

    setForegroundColor(canvas.color.fg, canvas);
    setBackgroundColor(canvas.color.bg, canvas);
}

function initializeCanvas(canvas) {
    window.addEventListener('resize', function() { resizeCanvas(canvas); });
}

function resizeCanvas(canvas) {
    let charWidth = WIDTH * 8;
    let charHeight = HEIGHT * 8;

    let body = document.getElementById('body');

    let xw = flp2(Math.floor(body.clientWidth / charWidth));
    let yw = flp2(Math.floor(body.clientHeight / charHeight));

    let w = Math.min(xw, yw);

    canvas.element.width = w * charWidth;
    canvas.element.height = w * charHeight;
    canvas.scalingFactor = w;

    renderBuffer(canvas);
}

// Round down to nearest power of 2.
function flp2(x) {
    x = x | (x >> 1);
    x = x | (x >> 2);
    x = x | (x >> 4);
    x = x | (x >> 8);
    x = x | (x >> 16);
    return x - (x >> 1);
}

function wrapCursor(canvas) {
    if (canvas.cursorPosition.col < 0) {
        canvas.cursorPosition.col = WIDTH-1;
        canvas.cursorPosition.row--;
    }

    if (canvas.cursorPosition.row < 0) {
        canvas.cursorPosition.row = 0;
    }

    if (canvas.cursorPosition.col >= WIDTH) {
        canvas.cursorPosition.col = 0;
        canvas.cursorPosition.row++;
    }

    if (canvas.cursorPosition.row >= HEIGHT) {
        canvas.cursorPosition.row = HEIGHT-1;
    }
}

function initializeWebSocket(canvas) {
    let origin = location.origin.substring(7);
    let socket = new WebSocket('ws://' + origin + '/ws');
    socket.binaryType = 'arraybuffer';

    socket.onopen = function(event) {
        let body = document.getElementById('body');
        body.addEventListener('keydown', function(event) {
            console.log(event);

            switch(event.key) {
                case "Enter":
                    canvas.cursorPosition.col = 0;
                    canvas.cursorPosition.row++;
                    break;
                case "Backspace":
                    canvas.cursorPosition.col--;
                    wrapCursor(canvas);
                    let msg = new Uint8Array(6);
                    msg[0] = 0x10; // KeyPress
                    msg[1] = 32;
                    msg[2] = canvas.cursorPosition.row;
                    msg[3] = canvas.cursorPosition.col;
                    msg[4] = canvas.color.fg;
                    msg[5] = canvas.color.bg;
                    socket.send(msg.buffer);
                    break;
                case "ArrowUp":
                    canvas.cursorPosition.row--;
                    break;
                case "ArrowDown":
                    canvas.cursorPosition.row++;
                    break;
                case "ArrowLeft":
                    canvas.cursorPosition.col--;
                    break;
                case "ArrowRight":
                    canvas.cursorPosition.col++;
                    break;
                default:
                    if (event.key.length === 1) {
                        let msg = new Uint8Array(6);
                        msg[0] = 0x10; // KeyPress
                        msg[1] = event.key.charCodeAt(0);
                        msg[2] = canvas.cursorPosition.row;
                        msg[3] = canvas.cursorPosition.col;
                        msg[4] = canvas.color.fg;
                        msg[5] = canvas.color.bg;
                        socket.send(msg.buffer);
                        canvas.cursorPosition.col++;
                    }
            }
            wrapCursor(canvas);
            renderBuffer(canvas);
        });
    };

    socket.onclose = function(event) {
        console.log('Closing socket.');
    };

    socket.onerror = function(event) {
        console.log('Error!');
        console.log(event);
    };

    socket.onmessage = function(event) {
        let data = new Uint8Array(event.data);

        switch(data[0]) {
            case 0x01:
                let row, col, i = 1;
                for (row = 0; row < HEIGHT; ++row) {
                    for (col = 0; col < WIDTH; ++col) {
                        canvas.buffer[row][col].char = data[i];
                        canvas.buffer[row][col].fg = data[i+1];
                        canvas.buffer[row][col].bg = data[i+2];
                        canvas.buffer[row][col].r = data[i+3];
                        i += 4;
                    }
                }
                renderBuffer(canvas);
                break;
        }
    };
}

function renderBuffer(canvas) {
    let ctx = canvas.element.getContext('2d');
    ctx.imageSmoothingEnabled = false;
    let pos;
    for (let row = 0; row < HEIGHT; row++) {
        for (let col = 0; col < WIDTH; col++) {
            pos = canvas.buffer[row][col];
            if (col === canvas.cursorPosition.col && row === canvas.cursorPosition.row) {
                let fg, bg;
                if (canvas.r) {
                    fg = pos.bg;
                    bg = pos.fg;
                } else {
                    fg = pos.fg;
                    bg = pos.bg;
                }
                if (fg === bg) {
                    fg = 1;
                    bg = 0;
                }
                putChar(pos.char, col, row, bg, fg, canvas, ctx);
            } else {
                let fg, bg;
                if (canvas.r) {
                    fg = pos.bg;
                    bg = pos.fg;
                } else {
                    fg = pos.fg;
                    bg = pos.bg;
                }
                putChar(pos.char, col, row, pos.fg, pos.bg, canvas, ctx);
            }
        }
    }
}

function putChar(char, cx, cy, fg, bg, canvas, ctx) {
    let y = (bg + fg * 16) * 8;
    let x = char * 8;
    ctx.drawImage(canvas.font,
        x, y, 8, 8,
        cx * 8 * canvas.scalingFactor, cy * 8 * canvas.scalingFactor,
        8 * canvas.scalingFactor, 8 * canvas.scalingFactor);
}

function setForegroundColor(colorIndex, canvas) {
    canvas.color.fg = colorIndex % C64Colors.length;
    for (let i = 0; i < C64Colors.length; ++i) {
        let el = document.getElementById('fg' + i);
        if (i === colorIndex) {
            el.classList.add('selected');
        } else {
            el.classList.remove('selected');
        }
    }
}

function setBackgroundColor(colorIndex, canvas) {
    canvas.color.bg = colorIndex % C64Colors.length;
    for (let i = 0; i < C64Colors.length; ++i) {
        let el = document.getElementById('bg' + i);
        if (i === colorIndex) {
            el.classList.add('selected');
        } else {
            el.classList.remove('selected');
        }
    }
}
