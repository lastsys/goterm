const WIDTH = 40;
const HEIGHT = 25;

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
                    a[i][j] = {char: ' ', bg: 0, fg: 1}
                }
            }
            return a;
        })(),
        cursorPosition: {row: 0, col: 0},
        charIndex: {
            '@': 0,
            'A': 1,
            'B': 2,
            'C': 3,
            'D': 4,
            'E': 5,
            'F': 6,
            'G': 7,
            'H': 8,
            'I': 9,
            'J': 10,
            'K': 11,
            'L': 12,
            'M': 13,
            'N': 14,
            'O': 15,
            'P': 16,
            'Q': 17,
            'R': 18,
            'S': 19,
            'T': 20,
            'U': 21,
            'V': 22,
            'W': 23,
            'X': 24,
            'Y': 25,
            'Z': 26,
            '[': 27,
            '£': 28,
            ']': 29,
            ' ': 32,
            '' : 32,
            '!': 33,
            '"': 34,
            '#': 35,
            '$': 36,
            '%': 37,
            '&': 38,
            "'": 39,
            '(': 40,
            ')': 41,
            '*': 42,
            '+': 43,
            ',': 44,
            '-': 45,
            '.': 46,
            '/': 47,
            '0': 48,
            '1': 49,
            '2': 50,
            '3': 51,
            '4': 52,
            '5': 53,
            '6': 54,
            '7': 55,
            '8': 56,
            '9': 57,
            ':': 58,
            ';': 59,
            '<': 60,
            '=': 61,
            '>': 62,
            '?': 63,
            'a': 129,
            'b': 130,
            'c': 131,
            'd': 132,
            'e': 133,
            'f': 134,
            'g': 135,
            'h': 136,
            'i': 137,
            'j': 138,
            'k': 139,
            'l': 140,
            'm': 141,
            'n': 142,
            'o': 143,
            'p': 144,
            'q': 145,
            'r': 146,
            's': 147,
            't': 148,
            'u': 149,
            'v': 150,
            'w': 151,
            'x': 152,
            'y': 153,
            'z': 154
        }
    };

    initializeCanvas(canvas);
    resizeCanvas(canvas);
    initializeWebSocket(canvas);
}

function loadFont() {
    let font = new Image();
    font.src = '/static/font.png';
    return font
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

function initializeWebSocket(canvas) {
    let socket = new WebSocket('ws://localhost:9000/ws');

    socket.onopen = function(event) {
        socket.send('Client is alive!');
        let body = document.getElementById('body');
        body.addEventListener('keypress', function(event) {
            socket.send(JSON.stringify({
                msg: 'key',
                key: event.key
            }));
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
        console.log('Got message!');
        let data = JSON.parse(event.data);
        canvas.buffer = data.chars;
        canvas.cursorPosition = data.cursorPosition;
        renderBuffer(canvas);
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
                putChar(' ', col, row, 1, 1, canvas, ctx);
            } else {
                putChar(pos.char, col, row, pos.fg, pos.bg, canvas, ctx);
            }
        }
    }
}

function putChar(char, cx, cy, fg, bg, canvas, ctx) {
    let y = (bg + fg * 16) * 8;
    let x = canvas.charIndex[char] * 8;
    ctx.drawImage(canvas.font,
        x, y, 8, 8,
        cx * 8 * canvas.scalingFactor, cy * 8 * canvas.scalingFactor,
        8 * canvas.scalingFactor, 8 * canvas.scalingFactor);
}
