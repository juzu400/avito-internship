import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10,
    duration: '30s',
};

const BASE_URL = 'http://localhost:8080';

export default function () {
    const payload = JSON.stringify({
        pull_request_id: `k6-pr-${__VU}-${Date.now()}`,
        pull_request_name: `k6-load-${__VU}-${Date.now()}`,
        author_id: 'u1',
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(`${BASE_URL}/pullRequest/create`, payload, params);

    check(res, {
        'create PR status is 201 or 200': (r) => r.status === 201 || r.status === 200,
    });

    sleep(0.1);
}
