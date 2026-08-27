import http from "k6/http";
import { check } from "k6";

export const options = {
    vus: 200,
    iterations: 100000,

    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ["p(95)<50"],
    },
};

http.setResponseCallback(http.expectedStatuses({ min: 200, max: 300 }, 404));

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
    const get = http.get(`${BASE_URL}/healthz`);

    check(get, {
        "GET /health returns 200": (r) => r.status === 200,
    });
}
