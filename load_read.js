import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 20,
    duration: '30s',
};

const BASE_URL = 'http://localhost:8080';

export default function () {
    const res1 = http.get(`${BASE_URL}/users/stats`);
    check(res1, {
        '/users/stats status is 200': (r) => r.status === 200,
    });

    const res2 = http.get(`${BASE_URL}/pullRequests/stats`);
    check(res2, {
        '/stats/pullRequests status is 200': (r) => r.status === 200,
    });

    sleep(0.1);
}
