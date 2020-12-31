# Benchmark Reports

This report shows the performance difference between Simple Feature's native
set operations (pure Go) and the corresponding GEOS set operations.

To re-run the reports, use the `run.sh` script (it generates the markdown
tables below).

**Operation:** Intersection

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 43.7µs | 48.9µs | 0.9 |
| 2<sup>3</sup> | 47.7µs | 54.6µs | 0.9 |
| 2<sup>4</sup> | 65.9µs | 59.1µs | 1.1 |
| 2<sup>5</sup> | 115µs | 74.8µs | 1.5 |
| 2<sup>6</sup> | 193µs | 95.3µs | 2.0 |
| 2<sup>7</sup> | 377µs | 145µs | 2.6 |
| 2<sup>8</sup> | 638µs | 214µs | 3.0 |
| 2<sup>9</sup> | 1.31ms | 389µs | 3.4 |
| 2<sup>10</sup> | 2.74ms | 719µs | 3.8 |
| 2<sup>11</sup> | 5.76ms | 1.7ms | 3.4 |
| 2<sup>12</sup> | 12.3ms | 2.82ms | 4.4 |
| 2<sup>13</sup> | 22.4ms | 5.73ms | 3.9 |
| 2<sup>14</sup> | 50.2ms | 11.4ms | 4.4 |

**Operation:** Union

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 40.6µs | 49.5µs | 0.8 |
| 2<sup>3</sup> | 51µs | 61.6µs | 0.8 |
| 2<sup>4</sup> | 72.9µs | 78.2µs | 0.9 |
| 2<sup>5</sup> | 125µs | 83.8µs | 1.5 |
| 2<sup>6</sup> | 211µs | 132µs | 1.6 |
| 2<sup>7</sup> | 401µs | 182µs | 2.2 |
| 2<sup>8</sup> | 754µs | 317µs | 2.4 |
| 2<sup>9</sup> | 1.41ms | 586µs | 2.4 |
| 2<sup>10</sup> | 2.78ms | 1.17ms | 2.4 |
| 2<sup>11</sup> | 5.98ms | 2.16ms | 2.8 |
| 2<sup>12</sup> | 12.1ms | 4.79ms | 2.5 |
| 2<sup>13</sup> | 28.3ms | 9.71ms | 2.9 |
| 2<sup>14</sup> | 56ms | 19.1ms | 2.9 |

**Operation:** Difference

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 47.6µs | 49.5µs | 1.0 |
| 2<sup>3</sup> | 48.6µs | 56.7µs | 0.9 |
| 2<sup>4</sup> | 72.1µs | 66.5µs | 1.1 |
| 2<sup>5</sup> | 120µs | 79.4µs | 1.5 |
| 2<sup>6</sup> | 219µs | 123µs | 1.8 |
| 2<sup>7</sup> | 367µs | 175µs | 2.1 |
| 2<sup>8</sup> | 1.17ms | 277µs | 4.2 |
| 2<sup>9</sup> | 1.34ms | 505µs | 2.7 |
| 2<sup>10</sup> | 2.83ms | 1.02ms | 2.8 |
| 2<sup>11</sup> | 6.58ms | 1.86ms | 3.5 |
| 2<sup>12</sup> | 11.9ms | 4.05ms | 2.9 |
| 2<sup>13</sup> | 25.8ms | 8.65ms | 3.0 |
| 2<sup>14</sup> | 56.7ms | 17.9ms | 3.2 |

**Operation:** SymmetricDifference

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 51.6µs | 67.6µs | 0.8 |
| 2<sup>3</sup> | 63.3µs | 81.9µs | 0.8 |
| 2<sup>4</sup> | 102µs | 106µs | 1.0 |
| 2<sup>5</sup> | 280µs | 134µs | 2.1 |
| 2<sup>6</sup> | 288µs | 210µs | 1.4 |
| 2<sup>7</sup> | 517µs | 353µs | 1.5 |
| 2<sup>8</sup> | 1.09ms | 666µs | 1.6 |
| 2<sup>9</sup> | 1.92ms | 1.15ms | 1.7 |
| 2<sup>10</sup> | 19.2ms | 2.35ms | 8.2 |
| 2<sup>11</sup> | 8.56ms | 4.47ms | 1.9 |
| 2<sup>12</sup> | 16.9ms | 9.93ms | 1.7 |
| 2<sup>13</sup> | 57.5ms | 20.4ms | 2.8 |
| 2<sup>14</sup> | 76.4ms | 37.8ms | 2.0 |
