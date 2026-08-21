import http from "k6/http";
import { check } from "k6";
import { jUnit } from 'https://jslib.k6.io/k6-summary/0.0.2/index.js';

export const options = {
    vus: 1,
    iterations: 1,

    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ["p(95)<50"],
    },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
    const key = "smoke-test";

    const put = http.put(
        `${BASE_URL}/v1/keys/${key}`,
        JSON.stringify({ value: "hello" }),
        {
            headers: {
                "Content-Type": "application/json",
            },
        },
    );

    check(put, {
        "PUT /v1/keys/{key} returns 2xx": (r) => r.status >= 200 && r.status < 300,
    });

    const get = http.get(`${BASE_URL}/v1/keys/${key}`);

    check(get, {
        "GET /v1/keys/{key} returns 200": (r) => r.status === 200,
    });

    const count = http.get(`${BASE_URL}/v1/count`);

    check(count, {
        "GET /v1/count returns 200": (r) => r.status === 200,
    });

    const del = http.del(`${BASE_URL}/v1/keys/${key}`);

    check(del, {
        "DELETE /v1/keys/{key} returns 2xx": (r) =>
            r.status >= 200 && r.status < 300,
    });
}

export function handleSummary(data) {
    return {
        'k6.smoke.junit.xml': jUnit(data),
    };
}
