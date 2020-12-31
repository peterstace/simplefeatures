# Benchmark Reports

This report shows the performance difference between Simple Feature's native
set operations (pure Go) and the corresponding GEOS set operations.

To re-run the reports, use the `run.sh` script (it generates the markdown
tables below).

The source code for the benchmarks below is [here](../perf). The benchmarks
create two regular polygons, each with `n` sides (where `n` is the input size
in the tables below). The polygons partially overlap with each other. The set
operation on the two regular polygons is what is actually timed.

**Operation:** Intersection

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 39.2µs | 46.7µs | 0.8 |
| 2<sup>3</sup> | 48.1µs | 54.3µs | 0.9 |
| 2<sup>4</sup> | 65.2µs | 59.5µs | 1.1 |
| 2<sup>5</sup> | 113µs | 72.1µs | 1.6 |
| 2<sup>6</sup> | 190µs | 94.9µs | 2.0 |
| 2<sup>7</sup> | 354µs | 144µs | 2.5 |
| 2<sup>8</sup> | 647µs | 215µs | 3.0 |
| 2<sup>9</sup> | 1.28ms | 385µs | 3.3 |
| 2<sup>10</sup> | 2.47ms | 718µs | 3.4 |
| 2<sup>11</sup> | 5.36ms | 1.46ms | 3.7 |
| 2<sup>12</sup> | 11.1ms | 2.66ms | 4.2 |
| 2<sup>13</sup> | 22.1ms | 5.73ms | 3.9 |
| 2<sup>14</sup> | 44.7ms | 11.6ms | 3.9 |

**Operation:** Union

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 40.8µs | 49.6µs | 0.8 |
| 2<sup>3</sup> | 50.2µs | 55.6µs | 0.9 |
| 2<sup>4</sup> | 72.3µs | 67.5µs | 1.1 |
| 2<sup>5</sup> | 122µs | 86µs | 1.4 |
| 2<sup>6</sup> | 215µs | 127µs | 1.7 |
| 2<sup>7</sup> | 390µs | 190µs | 2.1 |
| 2<sup>8</sup> | 729µs | 318µs | 2.3 |
| 2<sup>9</sup> | 1.42ms | 574µs | 2.5 |
| 2<sup>10</sup> | 2.83ms | 1.19ms | 2.4 |
| 2<sup>11</sup> | 5.99ms | 2.19ms | 2.7 |
| 2<sup>12</sup> | 12.7ms | 4.7ms | 2.7 |
| 2<sup>13</sup> | 25.9ms | 9.42ms | 2.7 |
| 2<sup>14</sup> | 55ms | 20.3ms | 2.7 |

**Operation:** Difference

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 40.1µs | 48.9µs | 0.8 |
| 2<sup>3</sup> | 48µs | 55.6µs | 0.9 |
| 2<sup>4</sup> | 68.7µs | 64.8µs | 1.1 |
| 2<sup>5</sup> | 117µs | 79.9µs | 1.5 |
| 2<sup>6</sup> | 203µs | 116µs | 1.7 |
| 2<sup>7</sup> | 370µs | 172µs | 2.1 |
| 2<sup>8</sup> | 691µs | 281µs | 2.5 |
| 2<sup>9</sup> | 1.37ms | 512µs | 2.7 |
| 2<sup>10</sup> | 2.66ms | 1.03ms | 2.6 |
| 2<sup>11</sup> | 5.63ms | 1.85ms | 3.0 |
| 2<sup>12</sup> | 12ms | 4.02ms | 3.0 |
| 2<sup>13</sup> | 25.1ms | 8.09ms | 3.1 |
| 2<sup>14</sup> | 53.2ms | 17.2ms | 3.1 |

**Operation:** SymmetricDifference

| Input Size | Simple Features | GEOS | Ratio |
| ---        | ---             | ---  | ---   |
| 2<sup>2</sup> | 51.5µs | 68.2µs | 0.8 |
| 2<sup>3</sup> | 63.9µs | 79.4µs | 0.8 |
| 2<sup>4</sup> | 98.8µs | 102µs | 1.0 |
| 2<sup>5</sup> | 161µs | 137µs | 1.2 |
| 2<sup>6</sup> | 286µs | 210µs | 1.4 |
| 2<sup>7</sup> | 512µs | 335µs | 1.5 |
| 2<sup>8</sup> | 1ms | 618µs | 1.6 |
| 2<sup>9</sup> | 1.91ms | 1.15ms | 1.7 |
| 2<sup>10</sup> | 3.93ms | 2.32ms | 1.7 |
| 2<sup>11</sup> | 8.06ms | 4.46ms | 1.8 |
| 2<sup>12</sup> | 17.1ms | 9.86ms | 1.7 |
| 2<sup>13</sup> | 34.1ms | 19.5ms | 1.7 |
| 2<sup>14</sup> | 71ms | 38.3ms | 1.9 |
