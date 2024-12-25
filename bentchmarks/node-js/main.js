const http = require('http');
const fs = require('fs');

const PORT = 3000;
const RESULTS_FILE = 'results.json';

// Pomocnicza funkcja do mnożenia macierzy 2D
function multiplyMatrices(mA, mB) {
    const rowsA = mA.length;
    const colsA = mA[0].length;
    const colsB = mB[0].length;
    const mC = new Array(rowsA).fill(0).map(() => new Array(colsB).fill(0));

    for (let i = 0; i < rowsA; i++) {
        for (let j = 0; j < colsB; j++) {
            for (let k = 0; k < colsA; k++) {
                mC[i][j] += mA[i][k] * mB[k][j];
            }
        }
    }
    return mC;
}

// Odbiór żądania POST
const requestListener = (req, res) => {
    if (req.method === 'POST') {
        let body = '';

        req.on('data', chunk => {
            body += chunk;
        });

        req.on('end', () => {
            try {
                const data = JSON.parse(body);
                const { taskName, matrixA, matrixB } = data;

                if (taskName === 'matrix_multiplication') {
                    const multiplicationResult = multiplyMatrices(matrixA, matrixB);

                    // Zapis do pliku
                    const resultObj = {
                        timestamp: new Date().toISOString(),
                        taskName: taskName,
                        matrixA,
                        matrixB,
                        result: multiplicationResult
                    };

                    // Dodaj wynik do pliku (dopisz jako kolejny element tablicy).
                    let resultsContent = [];
                    if (fs.existsSync(RESULTS_FILE)) {
                        const existing = fs.readFileSync(RESULTS_FILE, 'utf8');
                        if (existing.trim().length) {
                            resultsContent = JSON.parse(existing);
                        }
                    }
                    resultsContent.push(resultObj);
                    fs.writeFileSync(RESULTS_FILE, JSON.stringify(resultsContent, null, 2));

                    res.writeHead(200, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ status: 'ok', result: multiplicationResult }));
                } else {
                    res.writeHead(400, { 'Content-Type': 'application/json' });
                    res.end(JSON.stringify({ status: 'error', message: 'Unknown task' }));
                }
            } catch (error) {
                res.writeHead(500, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({ status: 'error', message: error.message }));
                console.log(error);
            }
        });
    } else {
        res.writeHead(404, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ status: 'not_found' }));
    }
};

const server = http.createServer(requestListener);
server.listen(PORT, () => {
    console.log(`Server is listening on port ${PORT}`);
});
