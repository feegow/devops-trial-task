import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 20 },
    { duration: '2m', target: 50 },
    { duration: '1m', target: 0 },
  ],
};

const host = __ENV.HOST || 'http://dev.local';

export default function () {
  http.get(`${host}/api/v1/orders`);
  http.get(`${host}/api/v2/orders`);
  sleep(0.2);
}