window.BENCHMARK_DATA = {
  "lastUpdate": 1772388694110,
  "repoUrl": "https://github.com/Baselyne-Systems/bulkhead",
  "entries": {
    "Go Benchmarks": [
      {
        "commit": {
          "author": {
            "email": "achyuth.1995@gmail.com",
            "name": "Achyuth Samudrala",
            "username": "achyuthnsamudrala"
          },
          "committer": {
            "email": "achyuth.1995@gmail.com",
            "name": "Achyuth Samudrala",
            "username": "achyuthnsamudrala"
          },
          "distinct": true,
          "id": "087ec37554e92b6885d344d893a23e4628fb4d76",
          "message": "Include e2e tests in coverage and exclude infra-only packages\n\nE2e tests cover gRPC handlers and cross-service paths that unit tests\ncan't reach. Exclude config, database, telemetry, testutil, and models\nfrom the coverage denominator since they're infra bootstrapping with\nno unit tests. Coverage goes from ~17% to ~65%.\n\nCo-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>",
          "timestamp": "2026-03-01T17:36:08Z",
          "tree_id": "a8369c16a569fdb65c738f79a920115e5c16a331",
          "url": "https://github.com/Baselyne-Systems/bulkhead/commit/087ec37554e92b6885d344d893a23e4628fb4d76"
        },
        "date": 1772387540809,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkRecordAction",
            "value": 905.6,
            "unit": "ns/op\t     675 B/op\t       5 allocs/op",
            "extra": "1226910 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - ns/op",
            "value": 905.6,
            "unit": "ns/op",
            "extra": "1226910 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - B/op",
            "value": 675,
            "unit": "B/op",
            "extra": "1226910 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1226910 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload",
            "value": 907,
            "unit": "ns/op\t     621 B/op\t       4 allocs/op",
            "extra": "1303317 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - ns/op",
            "value": 907,
            "unit": "ns/op",
            "extra": "1303317 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - B/op",
            "value": 621,
            "unit": "B/op",
            "extra": "1303317 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1303317 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction",
            "value": 83.08,
            "unit": "ns/op\t     240 B/op\t       1 allocs/op",
            "extra": "13718640 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - ns/op",
            "value": 83.08,
            "unit": "ns/op",
            "extra": "13718640 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - B/op",
            "value": 240,
            "unit": "B/op",
            "extra": "13718640 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "13718640 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100",
            "value": 4783,
            "unit": "ns/op\t    8200 B/op\t       8 allocs/op",
            "extra": "263112 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - ns/op",
            "value": 4783,
            "unit": "ns/op",
            "extra": "263112 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - B/op",
            "value": 8200,
            "unit": "B/op",
            "extra": "263112 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "263112 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000",
            "value": 51567,
            "unit": "ns/op\t   65545 B/op\t      11 allocs/op",
            "extra": "23266 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - ns/op",
            "value": 51567,
            "unit": "ns/op",
            "extra": "23266 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - B/op",
            "value": 65545,
            "unit": "B/op",
            "extra": "23266 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "23266 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000",
            "value": 911999,
            "unit": "ns/op\t  786448 B/op\t      15 allocs/op",
            "extra": "1296 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - ns/op",
            "value": 911999,
            "unit": "ns/op",
            "extra": "1296 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - B/op",
            "value": 786448,
            "unit": "B/op",
            "extra": "1296 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "1296 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter",
            "value": 26530,
            "unit": "ns/op\t   16392 B/op\t       9 allocs/op",
            "extra": "45049 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - ns/op",
            "value": 26530,
            "unit": "ns/op",
            "extra": "45049 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - B/op",
            "value": 16392,
            "unit": "B/op",
            "extra": "45049 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "45049 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost",
            "value": 837.1,
            "unit": "ns/op\t     472 B/op\t       5 allocs/op",
            "extra": "1391290 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - ns/op",
            "value": 837.1,
            "unit": "ns/op",
            "extra": "1391290 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - B/op",
            "value": 472,
            "unit": "B/op",
            "extra": "1391290 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1391290 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10",
            "value": 1882,
            "unit": "ns/op\t    5320 B/op\t       8 allocs/op",
            "extra": "600448 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - ns/op",
            "value": 1882,
            "unit": "ns/op",
            "extra": "600448 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - B/op",
            "value": 5320,
            "unit": "B/op",
            "extra": "600448 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "600448 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100",
            "value": 24566,
            "unit": "ns/op\t   43336 B/op\t      11 allocs/op",
            "extra": "48708 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - ns/op",
            "value": 24566,
            "unit": "ns/op",
            "extra": "48708 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - B/op",
            "value": 43336,
            "unit": "B/op",
            "extra": "48708 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "48708 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000",
            "value": 321643,
            "unit": "ns/op\t  346446 B/op\t      14 allocs/op",
            "extra": "3771 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - ns/op",
            "value": 321643,
            "unit": "ns/op",
            "extra": "3771 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - B/op",
            "value": 346446,
            "unit": "B/op",
            "extra": "3771 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - allocs/op",
            "value": 14,
            "unit": "allocs/op",
            "extra": "3771 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost",
            "value": 182.7,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "6501392 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - ns/op",
            "value": 182.7,
            "unit": "ns/op",
            "extra": "6501392 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "6501392 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "6501392 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage",
            "value": 454.1,
            "unit": "ns/op\t     388 B/op\t       3 allocs/op",
            "extra": "2483238 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - ns/op",
            "value": 454.1,
            "unit": "ns/op",
            "extra": "2483238 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - B/op",
            "value": 388,
            "unit": "B/op",
            "extra": "2483238 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2483238 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget",
            "value": 65.28,
            "unit": "ns/op\t     160 B/op\t       1 allocs/op",
            "extra": "17901151 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - ns/op",
            "value": 65.28,
            "unit": "ns/op",
            "extra": "17901151 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - B/op",
            "value": 160,
            "unit": "B/op",
            "extra": "17901151 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "17901151 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget",
            "value": 504,
            "unit": "ns/op\t     536 B/op\t       5 allocs/op",
            "extra": "2227755 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - ns/op",
            "value": 504,
            "unit": "ns/op",
            "extra": "2227755 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - B/op",
            "value": 536,
            "unit": "B/op",
            "extra": "2227755 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2227755 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget",
            "value": 98.27,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "12034668 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - ns/op",
            "value": 98.27,
            "unit": "ns/op",
            "extra": "12034668 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12034668 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "12034668 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate",
            "value": 434.2,
            "unit": "ns/op\t     385 B/op\t       3 allocs/op",
            "extra": "2667646 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - ns/op",
            "value": 434.2,
            "unit": "ns/op",
            "extra": "2667646 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - B/op",
            "value": 385,
            "unit": "B/op",
            "extra": "2667646 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2667646 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel",
            "value": 320.4,
            "unit": "ns/op\t     493 B/op\t       4 allocs/op",
            "extra": "3663442 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - ns/op",
            "value": 320.4,
            "unit": "ns/op",
            "extra": "3663442 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - B/op",
            "value": 493,
            "unit": "B/op",
            "extra": "3663442 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3663442 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel",
            "value": 76.27,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "14740102 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - ns/op",
            "value": 76.27,
            "unit": "ns/op",
            "extra": "14740102 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "14740102 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "14740102 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit",
            "value": 95.02,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "12401618 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - ns/op",
            "value": 95.02,
            "unit": "ns/op",
            "extra": "12401618 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12401618 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "12401618 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt",
            "value": 96.59,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "12482055 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - ns/op",
            "value": 96.59,
            "unit": "ns/op",
            "extra": "12482055 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12482055 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "12482055 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn",
            "value": 97.88,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11381451 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - ns/op",
            "value": 97.88,
            "unit": "ns/op",
            "extra": "11381451 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11381451 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11381451 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget",
            "value": 33.92,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "35969048 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - ns/op",
            "value": 33.92,
            "unit": "ns/op",
            "extra": "35969048 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "35969048 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "35969048 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert",
            "value": 470.4,
            "unit": "ns/op\t     536 B/op\t       5 allocs/op",
            "extra": "2563938 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - ns/op",
            "value": 470.4,
            "unit": "ns/op",
            "extra": "2563938 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - B/op",
            "value": 536,
            "unit": "B/op",
            "extra": "2563938 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2563938 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel",
            "value": 340.2,
            "unit": "ns/op\t     507 B/op\t       4 allocs/op",
            "extra": "3242878 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - ns/op",
            "value": 340.2,
            "unit": "ns/op",
            "extra": "3242878 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - B/op",
            "value": 507,
            "unit": "B/op",
            "extra": "3242878 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3242878 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes",
            "value": 25034,
            "unit": "ns/op\t    1816 B/op\t      19 allocs/op",
            "extra": "47404 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - ns/op",
            "value": 25034,
            "unit": "ns/op",
            "extra": "47404 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - B/op",
            "value": 1816,
            "unit": "B/op",
            "extra": "47404 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - allocs/op",
            "value": 19,
            "unit": "allocs/op",
            "extra": "47404 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset",
            "value": 190108,
            "unit": "ns/op\t     368 B/op\t       7 allocs/op",
            "extra": "6336 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - ns/op",
            "value": 190108,
            "unit": "ns/op",
            "extra": "6336 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "6336 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "6336 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation",
            "value": 82.36,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "14414301 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - ns/op",
            "value": 82.36,
            "unit": "ns/op",
            "extra": "14414301 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "14414301 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "14414301 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency",
            "value": 511.8,
            "unit": "ns/op\t     506 B/op\t       5 allocs/op",
            "extra": "2821822 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - ns/op",
            "value": 511.8,
            "unit": "ns/op",
            "extra": "2821822 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - B/op",
            "value": 506,
            "unit": "B/op",
            "extra": "2821822 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2821822 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold",
            "value": 96.73,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "12188739 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - ns/op",
            "value": 96.73,
            "unit": "ns/op",
            "extra": "12188739 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12188739 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "12188739 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload",
            "value": 2742,
            "unit": "ns/op\t      32 B/op\t       2 allocs/op",
            "extra": "453549 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - ns/op",
            "value": 2742,
            "unit": "ns/op",
            "extra": "453549 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - B/op",
            "value": 32,
            "unit": "B/op",
            "extra": "453549 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "453549 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload",
            "value": 7632095,
            "unit": "ns/op\t   49219 B/op\t       2 allocs/op",
            "extra": "157 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload - ns/op",
            "value": 7632095,
            "unit": "ns/op",
            "extra": "157 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload - B/op",
            "value": 49219,
            "unit": "B/op",
            "extra": "157 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "157 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch",
            "value": 194592,
            "unit": "ns/op\t    1414 B/op\t       1 allocs/op",
            "extra": "6237 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - ns/op",
            "value": 194592,
            "unit": "ns/op",
            "extra": "6237 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - B/op",
            "value": 1414,
            "unit": "B/op",
            "extra": "6237 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6237 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy",
            "value": 24.4,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "50942743 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - ns/op",
            "value": 24.4,
            "unit": "ns/op",
            "extra": "50942743 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "50942743 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "50942743 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel",
            "value": 4691,
            "unit": "ns/op\t     128 B/op\t       3 allocs/op",
            "extra": "256909 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - ns/op",
            "value": 4691,
            "unit": "ns/op",
            "extra": "256909 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - B/op",
            "value": 128,
            "unit": "B/op",
            "extra": "256909 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "256909 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B",
            "value": 5005,
            "unit": "ns/op\t      48 B/op\t       2 allocs/op",
            "extra": "240829 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - ns/op",
            "value": 5005,
            "unit": "ns/op",
            "extra": "240829 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "240829 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "240829 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B",
            "value": 16160,
            "unit": "ns/op\t     128 B/op\t       2 allocs/op",
            "extra": "75068 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - ns/op",
            "value": 16160,
            "unit": "ns/op",
            "extra": "75068 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - B/op",
            "value": 128,
            "unit": "B/op",
            "extra": "75068 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "75068 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB",
            "value": 150664,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8138 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - ns/op",
            "value": 150664,
            "unit": "ns/op",
            "extra": "8138 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8138 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8138 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB",
            "value": 1562341,
            "unit": "ns/op\t   10373 B/op\t       2 allocs/op",
            "extra": "772 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - ns/op",
            "value": 1562341,
            "unit": "ns/op",
            "extra": "772 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - B/op",
            "value": 10373,
            "unit": "B/op",
            "extra": "772 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "772 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB",
            "value": 16143095,
            "unit": "ns/op\t  106624 B/op\t       2 allocs/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - ns/op",
            "value": 16143095,
            "unit": "ns/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - B/op",
            "value": 106624,
            "unit": "B/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB",
            "value": 165434551,
            "unit": "ns/op\t 1049776 B/op\t       8 allocs/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB - ns/op",
            "value": 165434551,
            "unit": "ns/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB - B/op",
            "value": 1049776,
            "unit": "B/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card",
            "value": 148514,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8156 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - ns/op",
            "value": 148514,
            "unit": "ns/op",
            "extra": "8156 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8156 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8156 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key",
            "value": 148961,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8118 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - ns/op",
            "value": 148961,
            "unit": "ns/op",
            "extra": "8118 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8118 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8118 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email",
            "value": 139317,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8690 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - ns/op",
            "value": 139317,
            "unit": "ns/op",
            "extra": "8690 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8690 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8690 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone",
            "value": 137615,
            "unit": "ns/op\t    1040 B/op\t       2 allocs/op",
            "extra": "8584 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - ns/op",
            "value": 137615,
            "unit": "ns/op",
            "extra": "8584 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - B/op",
            "value": 1040,
            "unit": "B/op",
            "extra": "8584 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8584 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn",
            "value": 149469,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "7992 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - ns/op",
            "value": 149469,
            "unit": "ns/op",
            "extra": "7992 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "7992 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7992 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns",
            "value": 12085,
            "unit": "ns/op\t     272 B/op\t       4 allocs/op",
            "extra": "100449 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - ns/op",
            "value": 12085,
            "unit": "ns/op",
            "extra": "100449 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - B/op",
            "value": 272,
            "unit": "B/op",
            "extra": "100449 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "100449 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns",
            "value": 159062,
            "unit": "ns/op\t    1024 B/op\t       1 allocs/op",
            "extra": "7502 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - ns/op",
            "value": 159062,
            "unit": "ns/op",
            "extra": "7502 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - B/op",
            "value": 1024,
            "unit": "B/op",
            "extra": "7502 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "7502 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns",
            "value": 233227,
            "unit": "ns/op\t   11045 B/op\t       4 allocs/op",
            "extra": "5032 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - ns/op",
            "value": 233227,
            "unit": "ns/op",
            "extra": "5032 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - B/op",
            "value": 11045,
            "unit": "B/op",
            "extra": "5032 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5032 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public",
            "value": 15.74,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "72524923 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - ns/op",
            "value": 15.74,
            "unit": "ns/op",
            "extra": "72524923 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "72524923 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "72524923 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal",
            "value": 15.7,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "76634802 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - ns/op",
            "value": 15.7,
            "unit": "ns/op",
            "extra": "76634802 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "76634802 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "76634802 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential",
            "value": 24.18,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "50172648 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - ns/op",
            "value": 24.18,
            "unit": "ns/op",
            "extra": "50172648 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "50172648 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "50172648 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted",
            "value": 15.22,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "81423206 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - ns/op",
            "value": 15.22,
            "unit": "ns/op",
            "extra": "81423206 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "81423206 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "81423206 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api",
            "value": 25.21,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "46671032 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - ns/op",
            "value": 25.21,
            "unit": "ns/op",
            "extra": "46671032 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "46671032 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "46671032 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage",
            "value": 26.57,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "45672324 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - ns/op",
            "value": 26.57,
            "unit": "ns/op",
            "extra": "45672324 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "45672324 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "45672324 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log",
            "value": 23.72,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "51478498 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - ns/op",
            "value": 23.72,
            "unit": "ns/op",
            "extra": "51478498 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "51478498 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "51478498 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api",
            "value": 27.91,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "43653249 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - ns/op",
            "value": 27.91,
            "unit": "ns/op",
            "extra": "43653249 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "43653249 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "43653249 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket",
            "value": 27.82,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "43929619 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - ns/op",
            "value": 27.82,
            "unit": "ns/op",
            "extra": "43929619 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "43929619 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "43929619 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service",
            "value": 28.72,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "41216911 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - ns/op",
            "value": 28.72,
            "unit": "ns/op",
            "extra": "41216911 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "41216911 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "41216911 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel",
            "value": 11.56,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel - ns/op",
            "value": 11.56,
            "unit": "ns/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel",
            "value": 2904,
            "unit": "ns/op\t      96 B/op\t       3 allocs/op",
            "extra": "407850 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - ns/op",
            "value": 2904,
            "unit": "ns/op",
            "extra": "407850 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - B/op",
            "value": 96,
            "unit": "B/op",
            "extra": "407850 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "407850 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload",
            "value": 16155172,
            "unit": "ns/op\t  106623 B/op\t       2 allocs/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - ns/op",
            "value": 16155172,
            "unit": "ns/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - B/op",
            "value": 106623,
            "unit": "B/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload",
            "value": 639714,
            "unit": "ns/op\t    4116 B/op\t       1 allocs/op",
            "extra": "1887 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - ns/op",
            "value": 639714,
            "unit": "ns/op",
            "extra": "1887 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - B/op",
            "value": 4116,
            "unit": "B/op",
            "extra": "1887 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1887 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload",
            "value": 264737,
            "unit": "ns/op\t    2736 B/op\t       3 allocs/op",
            "extra": "4604 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - ns/op",
            "value": 264737,
            "unit": "ns/op",
            "extra": "4604 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - B/op",
            "value": 2736,
            "unit": "B/op",
            "extra": "4604 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4604 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule",
            "value": 1266,
            "unit": "ns/op\t    1079 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule - ns/op",
            "value": 1266,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule - B/op",
            "value": 1079,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule",
            "value": 139.4,
            "unit": "ns/op\t     336 B/op\t       2 allocs/op",
            "extra": "8439016 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - ns/op",
            "value": 139.4,
            "unit": "ns/op",
            "extra": "8439016 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - B/op",
            "value": 336,
            "unit": "B/op",
            "extra": "8439016 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8439016 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100",
            "value": 37768,
            "unit": "ns/op\t   84105 B/op\t     111 allocs/op",
            "extra": "31382 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - ns/op",
            "value": 37768,
            "unit": "ns/op",
            "extra": "31382 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - B/op",
            "value": 84105,
            "unit": "B/op",
            "extra": "31382 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - allocs/op",
            "value": 111,
            "unit": "allocs/op",
            "extra": "31382 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000",
            "value": 589832,
            "unit": "ns/op\t 1020250 B/op\t    1015 allocs/op",
            "extra": "1767 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - ns/op",
            "value": 589832,
            "unit": "ns/op",
            "extra": "1767 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - B/op",
            "value": 1020250,
            "unit": "B/op",
            "extra": "1767 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - allocs/op",
            "value": 1015,
            "unit": "allocs/op",
            "extra": "1767 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000",
            "value": 13720041,
            "unit": "ns/op\t14993700 B/op\t   10026 allocs/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - ns/op",
            "value": 13720041,
            "unit": "ns/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - B/op",
            "value": 14993700,
            "unit": "B/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - allocs/op",
            "value": 10026,
            "unit": "allocs/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy",
            "value": 29278,
            "unit": "ns/op\t   40834 B/op\t     109 allocs/op",
            "extra": "41191 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - ns/op",
            "value": 29278,
            "unit": "ns/op",
            "extra": "41191 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - B/op",
            "value": 40834,
            "unit": "B/op",
            "extra": "41191 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - allocs/op",
            "value": 109,
            "unit": "allocs/op",
            "extra": "41191 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest",
            "value": 1585,
            "unit": "ns/op\t    1127 B/op\t      10 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest - ns/op",
            "value": 1585,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest - B/op",
            "value": 1127,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest",
            "value": 1529,
            "unit": "ns/op\t     976 B/op\t       7 allocs/op",
            "extra": "797389 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - ns/op",
            "value": 1529,
            "unit": "ns/op",
            "extra": "797389 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - B/op",
            "value": 976,
            "unit": "B/op",
            "extra": "797389 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "797389 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100",
            "value": 30961,
            "unit": "ns/op\t   78921 B/op\t      11 allocs/op",
            "extra": "38710 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - ns/op",
            "value": 30961,
            "unit": "ns/op",
            "extra": "38710 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - B/op",
            "value": 78921,
            "unit": "B/op",
            "extra": "38710 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "38710 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000",
            "value": 502573,
            "unit": "ns/op\t  922712 B/op\t      15 allocs/op",
            "extra": "2365 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - ns/op",
            "value": 502573,
            "unit": "ns/op",
            "extra": "2365 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - B/op",
            "value": 922712,
            "unit": "B/op",
            "extra": "2365 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "2365 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000",
            "value": 14272801,
            "unit": "ns/op\t13341860 B/op\t      26 allocs/op",
            "extra": "84 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - ns/op",
            "value": 14272801,
            "unit": "ns/op",
            "extra": "84 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - B/op",
            "value": 13341860,
            "unit": "B/op",
            "extra": "84 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - allocs/op",
            "value": 26,
            "unit": "allocs/op",
            "extra": "84 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken",
            "value": 440.7,
            "unit": "ns/op\t     320 B/op\t       5 allocs/op",
            "extra": "2724184 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - ns/op",
            "value": 440.7,
            "unit": "ns/op",
            "extra": "2724184 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2724184 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2724184 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken",
            "value": 240.3,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "4977669 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - ns/op",
            "value": 240.3,
            "unit": "ns/op",
            "extra": "4977669 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "4977669 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4977669 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent",
            "value": 1055,
            "unit": "ns/op\t     631 B/op\t       5 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - ns/op",
            "value": 1055,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - B/op",
            "value": 631,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent",
            "value": 128.2,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "7999792 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - ns/op",
            "value": 128.2,
            "unit": "ns/op",
            "extra": "7999792 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "7999792 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7999792 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential",
            "value": 1644,
            "unit": "ns/op\t    1046 B/op\t      12 allocs/op",
            "extra": "711783 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - ns/op",
            "value": 1644,
            "unit": "ns/op",
            "extra": "711783 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - B/op",
            "value": 1046,
            "unit": "B/op",
            "extra": "711783 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "711783 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100",
            "value": 27102,
            "unit": "ns/op\t   54665 B/op\t      11 allocs/op",
            "extra": "44146 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - ns/op",
            "value": 27102,
            "unit": "ns/op",
            "extra": "44146 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - B/op",
            "value": 54665,
            "unit": "B/op",
            "extra": "44146 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "44146 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000",
            "value": 412561,
            "unit": "ns/op\t  693652 B/op\t      15 allocs/op",
            "extra": "2905 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - ns/op",
            "value": 412561,
            "unit": "ns/op",
            "extra": "2905 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - B/op",
            "value": 693652,
            "unit": "B/op",
            "extra": "2905 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "2905 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000",
            "value": 10831276,
            "unit": "ns/op\t10556889 B/op\t      26 allocs/op",
            "extra": "106 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - ns/op",
            "value": 10831276,
            "unit": "ns/op",
            "extra": "106 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - B/op",
            "value": 10556889,
            "unit": "B/op",
            "extra": "106 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - allocs/op",
            "value": 26,
            "unit": "allocs/op",
            "extra": "106 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent",
            "value": 845.8,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "1416304 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - ns/op",
            "value": 845.8,
            "unit": "ns/op",
            "extra": "1416304 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "1416304 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1416304 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel",
            "value": 1618,
            "unit": "ns/op\t     981 B/op\t      11 allocs/op",
            "extra": "805246 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - ns/op",
            "value": 1618,
            "unit": "ns/op",
            "extra": "805246 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - B/op",
            "value": 981,
            "unit": "B/op",
            "extra": "805246 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "805246 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel",
            "value": 235.5,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "5219336 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - ns/op",
            "value": 235.5,
            "unit": "ns/op",
            "extra": "5219336 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "5219336 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "5219336 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel",
            "value": 2005,
            "unit": "ns/op\t    1062 B/op\t      12 allocs/op",
            "extra": "589615 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - ns/op",
            "value": 2005,
            "unit": "ns/op",
            "extra": "589615 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - B/op",
            "value": 1062,
            "unit": "B/op",
            "extra": "589615 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "589615 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite",
            "value": 563.8,
            "unit": "ns/op\t     380 B/op\t       3 allocs/op",
            "extra": "2289219 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - ns/op",
            "value": 563.8,
            "unit": "ns/op",
            "extra": "2289219 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - B/op",
            "value": 380,
            "unit": "B/op",
            "extra": "2289219 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2289219 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels",
            "value": 1015,
            "unit": "ns/op\t     613 B/op\t       5 allocs/op",
            "extra": "1201899 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels - ns/op",
            "value": 1015,
            "unit": "ns/op",
            "extra": "1201899 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels - B/op",
            "value": 613,
            "unit": "B/op",
            "extra": "1201899 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1201899 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities",
            "value": 2037,
            "unit": "ns/op\t    2446 B/op\t       7 allocs/op",
            "extra": "650311 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - ns/op",
            "value": 2037,
            "unit": "ns/op",
            "extra": "650311 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - B/op",
            "value": 2446,
            "unit": "B/op",
            "extra": "650311 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "650311 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings",
            "value": 1111,
            "unit": "ns/op\t     679 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings - ns/op",
            "value": 1111,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings - B/op",
            "value": 679,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads",
            "value": 246.2,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "4879074 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - ns/op",
            "value": 246.2,
            "unit": "ns/op",
            "extra": "4879074 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "4879074 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4879074 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation",
            "value": 209221,
            "unit": "ns/op\t  226714 B/op\t      13 allocs/op",
            "extra": "5397 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - ns/op",
            "value": 209221,
            "unit": "ns/op",
            "extra": "5397 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - B/op",
            "value": 226714,
            "unit": "B/op",
            "extra": "5397 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "5397 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle",
            "value": 5591,
            "unit": "ns/op\t    2992 B/op\t      31 allocs/op",
            "extra": "210274 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - ns/op",
            "value": 5591,
            "unit": "ns/op",
            "extra": "210274 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - B/op",
            "value": 2992,
            "unit": "B/op",
            "extra": "210274 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - allocs/op",
            "value": 31,
            "unit": "allocs/op",
            "extra": "210274 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn",
            "value": 1824,
            "unit": "ns/op\t    1131 B/op\t      14 allocs/op",
            "extra": "668175 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - ns/op",
            "value": 1824,
            "unit": "ns/op",
            "extra": "668175 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - B/op",
            "value": 1131,
            "unit": "B/op",
            "extra": "668175 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - allocs/op",
            "value": 14,
            "unit": "allocs/op",
            "extra": "668175 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel",
            "value": 217.8,
            "unit": "ns/op\t     320 B/op\t       5 allocs/op",
            "extra": "5754844 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - ns/op",
            "value": 217.8,
            "unit": "ns/op",
            "extra": "5754844 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5754844 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "5754844 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel",
            "value": 112.9,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "10909506 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - ns/op",
            "value": 112.9,
            "unit": "ns/op",
            "extra": "10909506 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "10909506 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "10909506 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess",
            "value": 1731,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "718131 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - ns/op",
            "value": 1731,
            "unit": "ns/op",
            "extra": "718131 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "718131 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "718131 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K",
            "value": 121008245,
            "unit": "ns/op\t121304552 B/op\t      37 allocs/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - ns/op",
            "value": 121008245,
            "unit": "ns/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - B/op",
            "value": 121304552,
            "unit": "B/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants",
            "value": 256,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "4781634 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - ns/op",
            "value": 256,
            "unit": "ns/op",
            "extra": "4781634 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "4781634 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4781634 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace",
            "value": 1438,
            "unit": "ns/op\t    1079 B/op\t       8 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace - ns/op",
            "value": 1438,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace - B/op",
            "value": 1079,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace",
            "value": 200.4,
            "unit": "ns/op\t     456 B/op\t       3 allocs/op",
            "extra": "6103826 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - ns/op",
            "value": 200.4,
            "unit": "ns/op",
            "extra": "6103826 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - B/op",
            "value": 456,
            "unit": "B/op",
            "extra": "6103826 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "6103826 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100",
            "value": 43635,
            "unit": "ns/op\t  103178 B/op\t     111 allocs/op",
            "extra": "27565 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - ns/op",
            "value": 43635,
            "unit": "ns/op",
            "extra": "27565 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - B/op",
            "value": 103178,
            "unit": "B/op",
            "extra": "27565 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - allocs/op",
            "value": 111,
            "unit": "allocs/op",
            "extra": "27565 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000",
            "value": 849607,
            "unit": "ns/op\t 1235947 B/op\t    1016 allocs/op",
            "extra": "1533 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - ns/op",
            "value": 849607,
            "unit": "ns/op",
            "extra": "1533 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - B/op",
            "value": 1235947,
            "unit": "B/op",
            "extra": "1533 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - allocs/op",
            "value": 1016,
            "unit": "allocs/op",
            "extra": "1533 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000",
            "value": 15836037,
            "unit": "ns/op\t18535334 B/op\t   10026 allocs/op",
            "extra": "99 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - ns/op",
            "value": 15836037,
            "unit": "ns/op",
            "extra": "99 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - B/op",
            "value": 18535334,
            "unit": "B/op",
            "extra": "99 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - allocs/op",
            "value": 10026,
            "unit": "allocs/op",
            "extra": "99 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace",
            "value": 266,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "4526924 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - ns/op",
            "value": 266,
            "unit": "ns/op",
            "extra": "4526924 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4526924 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4526924 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec",
            "value": 1484,
            "unit": "ns/op\t    1331 B/op\t       9 allocs/op",
            "extra": "745443 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - ns/op",
            "value": 1484,
            "unit": "ns/op",
            "extra": "745443 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - B/op",
            "value": 1331,
            "unit": "B/op",
            "extra": "745443 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "745443 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "achyuth.1995@gmail.com",
            "name": "Achyuth Samudrala",
            "username": "achyuthnsamudrala"
          },
          "committer": {
            "email": "achyuth.1995@gmail.com",
            "name": "Achyuth Samudrala",
            "username": "achyuthnsamudrala"
          },
          "distinct": true,
          "id": "c11cb82d968e41feb98309636d305d33b8e1d8ae",
          "message": "Add benchmark table view to gh-pages grouped by service\n\nGenerate an HTML table of all benchmarks grouped by service and publish\nit as index.html on gh-pages, amending the benchmark-action commit so\nthere's only one commit per CI run. The table includes service TOC,\niterations, ns/op, B/op, and allocs/op columns.\n\nCo-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>",
          "timestamp": "2026-03-01T17:56:51Z",
          "tree_id": "dc1009b04a8bb4663bc599ae750e044b07185173",
          "url": "https://github.com/Baselyne-Systems/bulkhead/commit/c11cb82d968e41feb98309636d305d33b8e1d8ae"
        },
        "date": 1772388693572,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkRecordAction",
            "value": 1082,
            "unit": "ns/op\t     695 B/op\t       5 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - ns/op",
            "value": 1082,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - B/op",
            "value": 695,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload",
            "value": 995.6,
            "unit": "ns/op\t     628 B/op\t       4 allocs/op",
            "extra": "1207868 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - ns/op",
            "value": 995.6,
            "unit": "ns/op",
            "extra": "1207868 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - B/op",
            "value": 628,
            "unit": "B/op",
            "extra": "1207868 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1207868 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction",
            "value": 81.01,
            "unit": "ns/op\t     240 B/op\t       1 allocs/op",
            "extra": "14462427 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - ns/op",
            "value": 81.01,
            "unit": "ns/op",
            "extra": "14462427 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - B/op",
            "value": 240,
            "unit": "B/op",
            "extra": "14462427 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "14462427 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100",
            "value": 4183,
            "unit": "ns/op\t    8200 B/op\t       8 allocs/op",
            "extra": "247930 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - ns/op",
            "value": 4183,
            "unit": "ns/op",
            "extra": "247930 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - B/op",
            "value": 8200,
            "unit": "B/op",
            "extra": "247930 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "247930 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000",
            "value": 50007,
            "unit": "ns/op\t   65545 B/op\t      11 allocs/op",
            "extra": "23944 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - ns/op",
            "value": 50007,
            "unit": "ns/op",
            "extra": "23944 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - B/op",
            "value": 65545,
            "unit": "B/op",
            "extra": "23944 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "23944 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000",
            "value": 917346,
            "unit": "ns/op\t  786449 B/op\t      15 allocs/op",
            "extra": "1258 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - ns/op",
            "value": 917346,
            "unit": "ns/op",
            "extra": "1258 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - B/op",
            "value": 786449,
            "unit": "B/op",
            "extra": "1258 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "1258 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter",
            "value": 26152,
            "unit": "ns/op\t   16392 B/op\t       9 allocs/op",
            "extra": "45422 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - ns/op",
            "value": 26152,
            "unit": "ns/op",
            "extra": "45422 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - B/op",
            "value": 16392,
            "unit": "B/op",
            "extra": "45422 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "45422 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost",
            "value": 913.4,
            "unit": "ns/op\t     484 B/op\t       5 allocs/op",
            "extra": "1211668 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - ns/op",
            "value": 913.4,
            "unit": "ns/op",
            "extra": "1211668 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - B/op",
            "value": 484,
            "unit": "B/op",
            "extra": "1211668 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1211668 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10",
            "value": 2128,
            "unit": "ns/op\t    5320 B/op\t       8 allocs/op",
            "extra": "513976 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - ns/op",
            "value": 2128,
            "unit": "ns/op",
            "extra": "513976 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - B/op",
            "value": 5320,
            "unit": "B/op",
            "extra": "513976 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "513976 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100",
            "value": 24818,
            "unit": "ns/op\t   43336 B/op\t      11 allocs/op",
            "extra": "48301 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - ns/op",
            "value": 24818,
            "unit": "ns/op",
            "extra": "48301 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - B/op",
            "value": 43336,
            "unit": "B/op",
            "extra": "48301 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "48301 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000",
            "value": 325986,
            "unit": "ns/op\t  346447 B/op\t      14 allocs/op",
            "extra": "3766 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - ns/op",
            "value": 325986,
            "unit": "ns/op",
            "extra": "3766 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - B/op",
            "value": 346447,
            "unit": "B/op",
            "extra": "3766 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - allocs/op",
            "value": 14,
            "unit": "allocs/op",
            "extra": "3766 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost",
            "value": 234.7,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "5081656 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - ns/op",
            "value": 234.7,
            "unit": "ns/op",
            "extra": "5081656 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "5081656 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "5081656 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage",
            "value": 448.4,
            "unit": "ns/op\t     384 B/op\t       3 allocs/op",
            "extra": "2698641 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - ns/op",
            "value": 448.4,
            "unit": "ns/op",
            "extra": "2698641 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - B/op",
            "value": 384,
            "unit": "B/op",
            "extra": "2698641 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2698641 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget",
            "value": 62.88,
            "unit": "ns/op\t     160 B/op\t       1 allocs/op",
            "extra": "18700926 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - ns/op",
            "value": 62.88,
            "unit": "ns/op",
            "extra": "18700926 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - B/op",
            "value": 160,
            "unit": "B/op",
            "extra": "18700926 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "18700926 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget",
            "value": 464.9,
            "unit": "ns/op\t     536 B/op\t       5 allocs/op",
            "extra": "2572032 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - ns/op",
            "value": 464.9,
            "unit": "ns/op",
            "extra": "2572032 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - B/op",
            "value": 536,
            "unit": "B/op",
            "extra": "2572032 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2572032 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget",
            "value": 95.44,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "12594162 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - ns/op",
            "value": 95.44,
            "unit": "ns/op",
            "extra": "12594162 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12594162 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "12594162 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate",
            "value": 468.7,
            "unit": "ns/op\t     385 B/op\t       3 allocs/op",
            "extra": "2648811 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - ns/op",
            "value": 468.7,
            "unit": "ns/op",
            "extra": "2648811 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - B/op",
            "value": 385,
            "unit": "B/op",
            "extra": "2648811 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2648811 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel",
            "value": 487.8,
            "unit": "ns/op\t     517 B/op\t       4 allocs/op",
            "extra": "3099580 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - ns/op",
            "value": 487.8,
            "unit": "ns/op",
            "extra": "3099580 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - B/op",
            "value": 517,
            "unit": "B/op",
            "extra": "3099580 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3099580 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel",
            "value": 82.52,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "14917586 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - ns/op",
            "value": 82.52,
            "unit": "ns/op",
            "extra": "14917586 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "14917586 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "14917586 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit",
            "value": 102.1,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11825023 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - ns/op",
            "value": 102.1,
            "unit": "ns/op",
            "extra": "11825023 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11825023 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11825023 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt",
            "value": 102.2,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11861468 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - ns/op",
            "value": 102.2,
            "unit": "ns/op",
            "extra": "11861468 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11861468 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11861468 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn",
            "value": 102.4,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11273500 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - ns/op",
            "value": 102.4,
            "unit": "ns/op",
            "extra": "11273500 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11273500 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11273500 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget",
            "value": 35.79,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "33920034 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - ns/op",
            "value": 35.79,
            "unit": "ns/op",
            "extra": "33920034 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "33920034 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "33920034 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert",
            "value": 491.4,
            "unit": "ns/op\t     536 B/op\t       5 allocs/op",
            "extra": "2445145 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - ns/op",
            "value": 491.4,
            "unit": "ns/op",
            "extra": "2445145 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - B/op",
            "value": 536,
            "unit": "B/op",
            "extra": "2445145 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2445145 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel",
            "value": 419.1,
            "unit": "ns/op\t     513 B/op\t       4 allocs/op",
            "extra": "3263522 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - ns/op",
            "value": 419.1,
            "unit": "ns/op",
            "extra": "3263522 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - B/op",
            "value": 513,
            "unit": "B/op",
            "extra": "3263522 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "3263522 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes",
            "value": 25607,
            "unit": "ns/op\t    1816 B/op\t      19 allocs/op",
            "extra": "46609 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - ns/op",
            "value": 25607,
            "unit": "ns/op",
            "extra": "46609 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - B/op",
            "value": 1816,
            "unit": "B/op",
            "extra": "46609 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - allocs/op",
            "value": 19,
            "unit": "allocs/op",
            "extra": "46609 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset",
            "value": 195113,
            "unit": "ns/op\t     368 B/op\t       7 allocs/op",
            "extra": "6332 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - ns/op",
            "value": 195113,
            "unit": "ns/op",
            "extra": "6332 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "6332 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "6332 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation",
            "value": 88.62,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "13879921 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - ns/op",
            "value": 88.62,
            "unit": "ns/op",
            "extra": "13879921 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "13879921 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "13879921 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency",
            "value": 505.1,
            "unit": "ns/op\t     526 B/op\t       5 allocs/op",
            "extra": "2449292 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - ns/op",
            "value": 505.1,
            "unit": "ns/op",
            "extra": "2449292 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - B/op",
            "value": 526,
            "unit": "B/op",
            "extra": "2449292 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2449292 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold",
            "value": 104.2,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11190661 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - ns/op",
            "value": 104.2,
            "unit": "ns/op",
            "extra": "11190661 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11190661 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11190661 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload",
            "value": 2804,
            "unit": "ns/op\t      32 B/op\t       2 allocs/op",
            "extra": "432406 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - ns/op",
            "value": 2804,
            "unit": "ns/op",
            "extra": "432406 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - B/op",
            "value": 32,
            "unit": "B/op",
            "extra": "432406 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "432406 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload",
            "value": 7825194,
            "unit": "ns/op\t   49221 B/op\t       2 allocs/op",
            "extra": "153 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload - ns/op",
            "value": 7825194,
            "unit": "ns/op",
            "extra": "153 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload - B/op",
            "value": 49221,
            "unit": "B/op",
            "extra": "153 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "153 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch",
            "value": 196219,
            "unit": "ns/op\t    1408 B/op\t       1 allocs/op",
            "extra": "6085 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - ns/op",
            "value": 196219,
            "unit": "ns/op",
            "extra": "6085 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - B/op",
            "value": 1408,
            "unit": "B/op",
            "extra": "6085 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6085 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy",
            "value": 23.62,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "50456971 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - ns/op",
            "value": 23.62,
            "unit": "ns/op",
            "extra": "50456971 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "50456971 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "50456971 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel",
            "value": 4776,
            "unit": "ns/op\t     128 B/op\t       3 allocs/op",
            "extra": "254960 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - ns/op",
            "value": 4776,
            "unit": "ns/op",
            "extra": "254960 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - B/op",
            "value": 128,
            "unit": "B/op",
            "extra": "254960 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "254960 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B",
            "value": 5141,
            "unit": "ns/op\t      48 B/op\t       2 allocs/op",
            "extra": "240903 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - ns/op",
            "value": 5141,
            "unit": "ns/op",
            "extra": "240903 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "240903 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "240903 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B",
            "value": 16385,
            "unit": "ns/op\t     129 B/op\t       2 allocs/op",
            "extra": "73741 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - ns/op",
            "value": 16385,
            "unit": "ns/op",
            "extra": "73741 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - B/op",
            "value": 129,
            "unit": "B/op",
            "extra": "73741 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "73741 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB",
            "value": 151830,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "7894 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - ns/op",
            "value": 151830,
            "unit": "ns/op",
            "extra": "7894 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "7894 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7894 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB",
            "value": 1559420,
            "unit": "ns/op\t   10325 B/op\t       2 allocs/op",
            "extra": "774 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - ns/op",
            "value": 1559420,
            "unit": "ns/op",
            "extra": "774 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - B/op",
            "value": 10325,
            "unit": "B/op",
            "extra": "774 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "774 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB",
            "value": 16217467,
            "unit": "ns/op\t  106625 B/op\t       2 allocs/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - ns/op",
            "value": 16217467,
            "unit": "ns/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - B/op",
            "value": 106625,
            "unit": "B/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB",
            "value": 166247401,
            "unit": "ns/op\t 1050712 B/op\t      12 allocs/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB - ns/op",
            "value": 166247401,
            "unit": "ns/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB - B/op",
            "value": 1050712,
            "unit": "B/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn",
            "value": 151922,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "7855 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - ns/op",
            "value": 151922,
            "unit": "ns/op",
            "extra": "7855 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "7855 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7855 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card",
            "value": 152217,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "7917 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - ns/op",
            "value": 152217,
            "unit": "ns/op",
            "extra": "7917 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "7917 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7917 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key",
            "value": 150997,
            "unit": "ns/op\t    1040 B/op\t       2 allocs/op",
            "extra": "7819 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - ns/op",
            "value": 150997,
            "unit": "ns/op",
            "extra": "7819 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - B/op",
            "value": 1040,
            "unit": "B/op",
            "extra": "7819 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7819 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email",
            "value": 140998,
            "unit": "ns/op\t    1048 B/op\t       2 allocs/op",
            "extra": "8607 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - ns/op",
            "value": 140998,
            "unit": "ns/op",
            "extra": "8607 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - B/op",
            "value": 1048,
            "unit": "B/op",
            "extra": "8607 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8607 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone",
            "value": 139378,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8440 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - ns/op",
            "value": 139378,
            "unit": "ns/op",
            "extra": "8440 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8440 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8440 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns",
            "value": 12342,
            "unit": "ns/op\t     273 B/op\t       4 allocs/op",
            "extra": "96871 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - ns/op",
            "value": 12342,
            "unit": "ns/op",
            "extra": "96871 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - B/op",
            "value": 273,
            "unit": "B/op",
            "extra": "96871 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "96871 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns",
            "value": 162811,
            "unit": "ns/op\t    1024 B/op\t       1 allocs/op",
            "extra": "7620 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - ns/op",
            "value": 162811,
            "unit": "ns/op",
            "extra": "7620 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - B/op",
            "value": 1024,
            "unit": "B/op",
            "extra": "7620 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "7620 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns",
            "value": 233094,
            "unit": "ns/op\t   11034 B/op\t       4 allocs/op",
            "extra": "5047 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - ns/op",
            "value": 233094,
            "unit": "ns/op",
            "extra": "5047 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - B/op",
            "value": 11034,
            "unit": "B/op",
            "extra": "5047 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5047 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public",
            "value": 15.43,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "75763717 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - ns/op",
            "value": 15.43,
            "unit": "ns/op",
            "extra": "75763717 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "75763717 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "75763717 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal",
            "value": 15.41,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "76002751 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - ns/op",
            "value": 15.41,
            "unit": "ns/op",
            "extra": "76002751 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "76002751 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "76002751 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential",
            "value": 23.41,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "50938202 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - ns/op",
            "value": 23.41,
            "unit": "ns/op",
            "extra": "50938202 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "50938202 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "50938202 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted",
            "value": 14.94,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "79574757 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - ns/op",
            "value": 14.94,
            "unit": "ns/op",
            "extra": "79574757 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "79574757 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "79574757 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api",
            "value": 24.24,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "49629482 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - ns/op",
            "value": 24.24,
            "unit": "ns/op",
            "extra": "49629482 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "49629482 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "49629482 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage",
            "value": 26.14,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "45347238 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - ns/op",
            "value": 26.14,
            "unit": "ns/op",
            "extra": "45347238 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "45347238 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "45347238 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log",
            "value": 22.92,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "52123443 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - ns/op",
            "value": 22.92,
            "unit": "ns/op",
            "extra": "52123443 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "52123443 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "52123443 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api",
            "value": 26.51,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "45630009 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - ns/op",
            "value": 26.51,
            "unit": "ns/op",
            "extra": "45630009 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "45630009 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "45630009 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket",
            "value": 27.32,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "43289844 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - ns/op",
            "value": 27.32,
            "unit": "ns/op",
            "extra": "43289844 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "43289844 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "43289844 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service",
            "value": 28.79,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "41999474 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - ns/op",
            "value": 28.79,
            "unit": "ns/op",
            "extra": "41999474 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "41999474 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "41999474 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel",
            "value": 11.56,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel - ns/op",
            "value": 11.56,
            "unit": "ns/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel",
            "value": 2934,
            "unit": "ns/op\t      96 B/op\t       3 allocs/op",
            "extra": "411512 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - ns/op",
            "value": 2934,
            "unit": "ns/op",
            "extra": "411512 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - B/op",
            "value": 96,
            "unit": "B/op",
            "extra": "411512 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "411512 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload",
            "value": 16177008,
            "unit": "ns/op\t  106622 B/op\t       2 allocs/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - ns/op",
            "value": 16177008,
            "unit": "ns/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - B/op",
            "value": 106622,
            "unit": "B/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "73 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload",
            "value": 649432,
            "unit": "ns/op\t    4116 B/op\t       1 allocs/op",
            "extra": "1880 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - ns/op",
            "value": 649432,
            "unit": "ns/op",
            "extra": "1880 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - B/op",
            "value": 4116,
            "unit": "B/op",
            "extra": "1880 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1880 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload",
            "value": 268416,
            "unit": "ns/op\t    2736 B/op\t       3 allocs/op",
            "extra": "4525 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - ns/op",
            "value": 268416,
            "unit": "ns/op",
            "extra": "4525 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - B/op",
            "value": 2736,
            "unit": "B/op",
            "extra": "4525 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4525 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule",
            "value": 1412,
            "unit": "ns/op\t    1079 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule - ns/op",
            "value": 1412,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule - B/op",
            "value": 1079,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule",
            "value": 153.1,
            "unit": "ns/op\t     336 B/op\t       2 allocs/op",
            "extra": "7251386 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - ns/op",
            "value": 153.1,
            "unit": "ns/op",
            "extra": "7251386 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - B/op",
            "value": 336,
            "unit": "B/op",
            "extra": "7251386 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7251386 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100",
            "value": 38376,
            "unit": "ns/op\t   84105 B/op\t     111 allocs/op",
            "extra": "31323 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - ns/op",
            "value": 38376,
            "unit": "ns/op",
            "extra": "31323 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - B/op",
            "value": 84105,
            "unit": "B/op",
            "extra": "31323 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - allocs/op",
            "value": 111,
            "unit": "allocs/op",
            "extra": "31323 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000",
            "value": 586108,
            "unit": "ns/op\t 1020249 B/op\t    1015 allocs/op",
            "extra": "2026 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - ns/op",
            "value": 586108,
            "unit": "ns/op",
            "extra": "2026 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - B/op",
            "value": 1020249,
            "unit": "B/op",
            "extra": "2026 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - allocs/op",
            "value": 1015,
            "unit": "allocs/op",
            "extra": "2026 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000",
            "value": 14627353,
            "unit": "ns/op\t14993700 B/op\t   10026 allocs/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - ns/op",
            "value": 14627353,
            "unit": "ns/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - B/op",
            "value": 14993700,
            "unit": "B/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - allocs/op",
            "value": 10026,
            "unit": "allocs/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy",
            "value": 30157,
            "unit": "ns/op\t   40845 B/op\t     109 allocs/op",
            "extra": "39717 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - ns/op",
            "value": 30157,
            "unit": "ns/op",
            "extra": "39717 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - B/op",
            "value": 40845,
            "unit": "B/op",
            "extra": "39717 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - allocs/op",
            "value": 109,
            "unit": "allocs/op",
            "extra": "39717 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest",
            "value": 1637,
            "unit": "ns/op\t    1128 B/op\t      10 allocs/op",
            "extra": "991371 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest - ns/op",
            "value": 1637,
            "unit": "ns/op",
            "extra": "991371 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest - B/op",
            "value": 1128,
            "unit": "B/op",
            "extra": "991371 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "991371 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest",
            "value": 1713,
            "unit": "ns/op\t     976 B/op\t       7 allocs/op",
            "extra": "648717 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - ns/op",
            "value": 1713,
            "unit": "ns/op",
            "extra": "648717 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - B/op",
            "value": 976,
            "unit": "B/op",
            "extra": "648717 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "648717 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100",
            "value": 31740,
            "unit": "ns/op\t   78921 B/op\t      11 allocs/op",
            "extra": "37290 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - ns/op",
            "value": 31740,
            "unit": "ns/op",
            "extra": "37290 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - B/op",
            "value": 78921,
            "unit": "B/op",
            "extra": "37290 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "37290 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000",
            "value": 555029,
            "unit": "ns/op\t  922715 B/op\t      15 allocs/op",
            "extra": "2361 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - ns/op",
            "value": 555029,
            "unit": "ns/op",
            "extra": "2361 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - B/op",
            "value": 922715,
            "unit": "B/op",
            "extra": "2361 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "2361 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000",
            "value": 14469129,
            "unit": "ns/op\t13341858 B/op\t      26 allocs/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - ns/op",
            "value": 14469129,
            "unit": "ns/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - B/op",
            "value": 13341858,
            "unit": "B/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - allocs/op",
            "value": 26,
            "unit": "allocs/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken",
            "value": 445.2,
            "unit": "ns/op\t     320 B/op\t       5 allocs/op",
            "extra": "2698558 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - ns/op",
            "value": 445.2,
            "unit": "ns/op",
            "extra": "2698558 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2698558 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2698558 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken",
            "value": 242.2,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "4939941 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - ns/op",
            "value": 242.2,
            "unit": "ns/op",
            "extra": "4939941 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "4939941 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4939941 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent",
            "value": 1084,
            "unit": "ns/op\t     631 B/op\t       5 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - ns/op",
            "value": 1084,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - B/op",
            "value": 631,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent",
            "value": 127.8,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "9201870 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - ns/op",
            "value": 127.8,
            "unit": "ns/op",
            "extra": "9201870 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "9201870 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9201870 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential",
            "value": 1724,
            "unit": "ns/op\t    1034 B/op\t      12 allocs/op",
            "extra": "850771 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - ns/op",
            "value": 1724,
            "unit": "ns/op",
            "extra": "850771 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - B/op",
            "value": 1034,
            "unit": "B/op",
            "extra": "850771 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "850771 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100",
            "value": 28425,
            "unit": "ns/op\t   54664 B/op\t      11 allocs/op",
            "extra": "41832 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - ns/op",
            "value": 28425,
            "unit": "ns/op",
            "extra": "41832 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - B/op",
            "value": 54664,
            "unit": "B/op",
            "extra": "41832 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "41832 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000",
            "value": 481128,
            "unit": "ns/op\t  693654 B/op\t      15 allocs/op",
            "extra": "2385 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - ns/op",
            "value": 481128,
            "unit": "ns/op",
            "extra": "2385 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - B/op",
            "value": 693654,
            "unit": "B/op",
            "extra": "2385 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "2385 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000",
            "value": 12050959,
            "unit": "ns/op\t10556883 B/op\t      25 allocs/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - ns/op",
            "value": 12050959,
            "unit": "ns/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - B/op",
            "value": 10556883,
            "unit": "B/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - allocs/op",
            "value": 25,
            "unit": "allocs/op",
            "extra": "100 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent",
            "value": 988.9,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "1212394 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - ns/op",
            "value": 988.9,
            "unit": "ns/op",
            "extra": "1212394 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "1212394 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1212394 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel",
            "value": 1654,
            "unit": "ns/op\t     987 B/op\t      11 allocs/op",
            "extra": "744686 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - ns/op",
            "value": 1654,
            "unit": "ns/op",
            "extra": "744686 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - B/op",
            "value": 987,
            "unit": "B/op",
            "extra": "744686 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "744686 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel",
            "value": 255.8,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "4432791 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - ns/op",
            "value": 255.8,
            "unit": "ns/op",
            "extra": "4432791 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "4432791 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4432791 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel",
            "value": 2193,
            "unit": "ns/op\t    1066 B/op\t      12 allocs/op",
            "extra": "569846 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - ns/op",
            "value": 2193,
            "unit": "ns/op",
            "extra": "569846 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - B/op",
            "value": 1066,
            "unit": "B/op",
            "extra": "569846 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "569846 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite",
            "value": 583.9,
            "unit": "ns/op\t     375 B/op\t       3 allocs/op",
            "extra": "2091376 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - ns/op",
            "value": 583.9,
            "unit": "ns/op",
            "extra": "2091376 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - B/op",
            "value": 375,
            "unit": "B/op",
            "extra": "2091376 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2091376 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels",
            "value": 1134,
            "unit": "ns/op\t     631 B/op\t       5 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels - ns/op",
            "value": 1134,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels - B/op",
            "value": 631,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities",
            "value": 2139,
            "unit": "ns/op\t    2447 B/op\t       7 allocs/op",
            "extra": "640065 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - ns/op",
            "value": 2139,
            "unit": "ns/op",
            "extra": "640065 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - B/op",
            "value": 2447,
            "unit": "B/op",
            "extra": "640065 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "640065 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings",
            "value": 1117,
            "unit": "ns/op\t     679 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings - ns/op",
            "value": 1117,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings - B/op",
            "value": 679,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads",
            "value": 257.9,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "4605174 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - ns/op",
            "value": 257.9,
            "unit": "ns/op",
            "extra": "4605174 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "4605174 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4605174 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation",
            "value": 215569,
            "unit": "ns/op\t  226713 B/op\t      13 allocs/op",
            "extra": "5043 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - ns/op",
            "value": 215569,
            "unit": "ns/op",
            "extra": "5043 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - B/op",
            "value": 226713,
            "unit": "B/op",
            "extra": "5043 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "5043 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle",
            "value": 6477,
            "unit": "ns/op\t    2992 B/op\t      31 allocs/op",
            "extra": "185853 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - ns/op",
            "value": 6477,
            "unit": "ns/op",
            "extra": "185853 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - B/op",
            "value": 2992,
            "unit": "B/op",
            "extra": "185853 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - allocs/op",
            "value": 31,
            "unit": "allocs/op",
            "extra": "185853 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn",
            "value": 1930,
            "unit": "ns/op\t    1137 B/op\t      14 allocs/op",
            "extra": "626056 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - ns/op",
            "value": 1930,
            "unit": "ns/op",
            "extra": "626056 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - B/op",
            "value": 1137,
            "unit": "B/op",
            "extra": "626056 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - allocs/op",
            "value": 14,
            "unit": "allocs/op",
            "extra": "626056 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel",
            "value": 215.8,
            "unit": "ns/op\t     320 B/op\t       5 allocs/op",
            "extra": "5643994 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - ns/op",
            "value": 215.8,
            "unit": "ns/op",
            "extra": "5643994 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5643994 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "5643994 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel",
            "value": 114.7,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "10664101 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - ns/op",
            "value": 114.7,
            "unit": "ns/op",
            "extra": "10664101 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "10664101 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "10664101 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess",
            "value": 1925,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "631057 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - ns/op",
            "value": 1925,
            "unit": "ns/op",
            "extra": "631057 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "631057 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "631057 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K",
            "value": 129093443,
            "unit": "ns/op\t121304552 B/op\t      37 allocs/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - ns/op",
            "value": 129093443,
            "unit": "ns/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - B/op",
            "value": 121304552,
            "unit": "B/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - allocs/op",
            "value": 37,
            "unit": "allocs/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants",
            "value": 270.3,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "4509303 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - ns/op",
            "value": 270.3,
            "unit": "ns/op",
            "extra": "4509303 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "4509303 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4509303 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace",
            "value": 1425,
            "unit": "ns/op\t    1079 B/op\t       8 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace - ns/op",
            "value": 1425,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace - B/op",
            "value": 1079,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace",
            "value": 190.1,
            "unit": "ns/op\t     456 B/op\t       3 allocs/op",
            "extra": "6224538 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - ns/op",
            "value": 190.1,
            "unit": "ns/op",
            "extra": "6224538 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - B/op",
            "value": 456,
            "unit": "B/op",
            "extra": "6224538 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "6224538 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100",
            "value": 45283,
            "unit": "ns/op\t  103178 B/op\t     111 allocs/op",
            "extra": "24004 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - ns/op",
            "value": 45283,
            "unit": "ns/op",
            "extra": "24004 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - B/op",
            "value": 103178,
            "unit": "B/op",
            "extra": "24004 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - allocs/op",
            "value": 111,
            "unit": "allocs/op",
            "extra": "24004 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000",
            "value": 818949,
            "unit": "ns/op\t 1235943 B/op\t    1016 allocs/op",
            "extra": "1544 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - ns/op",
            "value": 818949,
            "unit": "ns/op",
            "extra": "1544 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - B/op",
            "value": 1235943,
            "unit": "B/op",
            "extra": "1544 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - allocs/op",
            "value": 1016,
            "unit": "allocs/op",
            "extra": "1544 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000",
            "value": 15998275,
            "unit": "ns/op\t18535327 B/op\t   10026 allocs/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - ns/op",
            "value": 15998275,
            "unit": "ns/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - B/op",
            "value": 18535327,
            "unit": "B/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - allocs/op",
            "value": 10026,
            "unit": "allocs/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace",
            "value": 348.8,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "3435414 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - ns/op",
            "value": 348.8,
            "unit": "ns/op",
            "extra": "3435414 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3435414 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3435414 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec",
            "value": 1571,
            "unit": "ns/op\t    1328 B/op\t       9 allocs/op",
            "extra": "771390 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - ns/op",
            "value": 1571,
            "unit": "ns/op",
            "extra": "771390 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - B/op",
            "value": 1328,
            "unit": "B/op",
            "extra": "771390 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "771390 times\n4 procs"
          }
        ]
      }
    ]
  }
}