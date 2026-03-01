window.BENCHMARK_DATA = {
  "lastUpdate": 1772396861346,
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
          "id": "2684992d17697b24ab6a7e376146243784015c1a",
          "message": "Raise benchmark alert threshold to 150% to reduce CI noise\n\nCo-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>",
          "timestamp": "2026-03-01T20:00:57Z",
          "tree_id": "ef5b560e6934b17dda06d9fb6ae05c76b8edba4f",
          "url": "https://github.com/Baselyne-Systems/bulkhead/commit/2684992d17697b24ab6a7e376146243784015c1a"
        },
        "date": 1772396860436,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkRecordAction",
            "value": 906.9,
            "unit": "ns/op\t     675 B/op\t       5 allocs/op",
            "extra": "1220890 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - ns/op",
            "value": 906.9,
            "unit": "ns/op",
            "extra": "1220890 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - B/op",
            "value": 675,
            "unit": "B/op",
            "extra": "1220890 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1220890 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload",
            "value": 883.9,
            "unit": "ns/op\t     625 B/op\t       4 allocs/op",
            "extra": "1256510 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - ns/op",
            "value": 883.9,
            "unit": "ns/op",
            "extra": "1256510 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - B/op",
            "value": 625,
            "unit": "B/op",
            "extra": "1256510 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_LargePayload - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1256510 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction",
            "value": 80.24,
            "unit": "ns/op\t     240 B/op\t       1 allocs/op",
            "extra": "15114126 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - ns/op",
            "value": 80.24,
            "unit": "ns/op",
            "extra": "15114126 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - B/op",
            "value": 240,
            "unit": "B/op",
            "extra": "15114126 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "15114126 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100",
            "value": 4145,
            "unit": "ns/op\t    8200 B/op\t       8 allocs/op",
            "extra": "275905 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - ns/op",
            "value": 4145,
            "unit": "ns/op",
            "extra": "275905 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - B/op",
            "value": 8200,
            "unit": "B/op",
            "extra": "275905 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=100 - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "275905 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000",
            "value": 50247,
            "unit": "ns/op\t   65545 B/op\t      11 allocs/op",
            "extra": "24607 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - ns/op",
            "value": 50247,
            "unit": "ns/op",
            "extra": "24607 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - B/op",
            "value": 65545,
            "unit": "B/op",
            "extra": "24607 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=1000 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "24607 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000",
            "value": 936315,
            "unit": "ns/op\t  786450 B/op\t      15 allocs/op",
            "extra": "1111 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - ns/op",
            "value": 936315,
            "unit": "ns/op",
            "extra": "1111 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - B/op",
            "value": 786450,
            "unit": "B/op",
            "extra": "1111 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_Scaling/n=10000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "1111 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter",
            "value": 25567,
            "unit": "ns/op\t   16392 B/op\t       9 allocs/op",
            "extra": "46848 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - ns/op",
            "value": 25567,
            "unit": "ns/op",
            "extra": "46848 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - B/op",
            "value": 16392,
            "unit": "B/op",
            "extra": "46848 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_MultiFilter - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "46848 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_Parallel",
            "value": 1094,
            "unit": "ns/op\t     679 B/op\t       5 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_Parallel - ns/op",
            "value": 1094,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_Parallel - B/op",
            "value": 679,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_Parallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction_Parallel",
            "value": 103.9,
            "unit": "ns/op\t     240 B/op\t       1 allocs/op",
            "extra": "11466252 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction_Parallel - ns/op",
            "value": 103.9,
            "unit": "ns/op",
            "extra": "11466252 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction_Parallel - B/op",
            "value": 240,
            "unit": "B/op",
            "extra": "11466252 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAction_Parallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "11466252 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/100B",
            "value": 903.2,
            "unit": "ns/op\t     627 B/op\t       4 allocs/op",
            "extra": "1222940 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/100B - ns/op",
            "value": 903.2,
            "unit": "ns/op",
            "extra": "1222940 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/100B - B/op",
            "value": 627,
            "unit": "B/op",
            "extra": "1222940 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/100B - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1222940 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/1KB",
            "value": 915,
            "unit": "ns/op\t     625 B/op\t       4 allocs/op",
            "extra": "1243524 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/1KB - ns/op",
            "value": 915,
            "unit": "ns/op",
            "extra": "1243524 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/1KB - B/op",
            "value": 625,
            "unit": "B/op",
            "extra": "1243524 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/1KB - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1243524 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/4KB",
            "value": 906.7,
            "unit": "ns/op\t     625 B/op\t       4 allocs/op",
            "extra": "1254268 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/4KB - ns/op",
            "value": 906.7,
            "unit": "ns/op",
            "extra": "1254268 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/4KB - B/op",
            "value": 625,
            "unit": "B/op",
            "extra": "1254268 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/4KB - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1254268 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/16KB",
            "value": 917.2,
            "unit": "ns/op\t     624 B/op\t       4 allocs/op",
            "extra": "1268846 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/16KB - ns/op",
            "value": 917.2,
            "unit": "ns/op",
            "extra": "1268846 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/16KB - B/op",
            "value": 624,
            "unit": "B/op",
            "extra": "1268846 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/16KB - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1268846 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/64KB",
            "value": 914.8,
            "unit": "ns/op\t     625 B/op\t       4 allocs/op",
            "extra": "1244832 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/64KB - ns/op",
            "value": 914.8,
            "unit": "ns/op",
            "extra": "1244832 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/64KB - B/op",
            "value": 625,
            "unit": "B/op",
            "extra": "1244832 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_VaryingPayloadSizes/64KB - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1244832 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Allowed",
            "value": 913.8,
            "unit": "ns/op\t     622 B/op\t       4 allocs/op",
            "extra": "1288444 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Allowed - ns/op",
            "value": 913.8,
            "unit": "ns/op",
            "extra": "1288444 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Allowed - B/op",
            "value": 622,
            "unit": "B/op",
            "extra": "1288444 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Allowed - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1288444 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Denied",
            "value": 925.9,
            "unit": "ns/op\t     623 B/op\t       4 allocs/op",
            "extra": "1283054 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Denied - ns/op",
            "value": 925.9,
            "unit": "ns/op",
            "extra": "1283054 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Denied - B/op",
            "value": 623,
            "unit": "B/op",
            "extra": "1283054 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Denied - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1283054 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Escalated",
            "value": 909,
            "unit": "ns/op\t     627 B/op\t       4 allocs/op",
            "extra": "1220542 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Escalated - ns/op",
            "value": 909,
            "unit": "ns/op",
            "extra": "1220542 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Escalated - B/op",
            "value": 627,
            "unit": "B/op",
            "extra": "1220542 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Escalated - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1220542 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Error",
            "value": 920.5,
            "unit": "ns/op\t     628 B/op\t       4 allocs/op",
            "extra": "1202898 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Error - ns/op",
            "value": 920.5,
            "unit": "ns/op",
            "extra": "1202898 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Error - B/op",
            "value": 628,
            "unit": "B/op",
            "extra": "1202898 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_AllOutcomes/Error - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1202898 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_WithLatencies",
            "value": 1078,
            "unit": "ns/op\t     687 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_WithLatencies - ns/op",
            "value": 1078,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_WithLatencies - B/op",
            "value": 687,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_WithLatencies - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=100",
            "value": 4116,
            "unit": "ns/op\t    8200 B/op\t       8 allocs/op",
            "extra": "261777 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=100 - ns/op",
            "value": 4116,
            "unit": "ns/op",
            "extra": "261777 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=100 - B/op",
            "value": 8200,
            "unit": "B/op",
            "extra": "261777 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=100 - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "261777 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=1000",
            "value": 50812,
            "unit": "ns/op\t   65545 B/op\t      11 allocs/op",
            "extra": "23581 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=1000 - ns/op",
            "value": 50812,
            "unit": "ns/op",
            "extra": "23581 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=1000 - B/op",
            "value": 65545,
            "unit": "B/op",
            "extra": "23581 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=1000 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "23581 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=10000",
            "value": 960822,
            "unit": "ns/op\t  786453 B/op\t      15 allocs/op",
            "extra": "1272 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=10000 - ns/op",
            "value": 960822,
            "unit": "ns/op",
            "extra": "1272 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=10000 - B/op",
            "value": 786453,
            "unit": "B/op",
            "extra": "1272 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=10000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "1272 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=50000",
            "value": 5910630,
            "unit": "ns/op\t 4964372 B/op\t      20 allocs/op",
            "extra": "205 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=50000 - ns/op",
            "value": 5910630,
            "unit": "ns/op",
            "extra": "205 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=50000 - B/op",
            "value": 4964372,
            "unit": "B/op",
            "extra": "205 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_ScalingExtended/n=50000 - allocs/op",
            "value": 20,
            "unit": "allocs/op",
            "extra": "205 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_AllFilters",
            "value": 182769,
            "unit": "ns/op\t   65545 B/op\t      11 allocs/op",
            "extra": "5846 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_AllFilters - ns/op",
            "value": 182769,
            "unit": "ns/op",
            "extra": "5846 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_AllFilters - B/op",
            "value": 65545,
            "unit": "B/op",
            "extra": "5846 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_AllFilters - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "5846 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_TimeRange",
            "value": 5241956,
            "unit": "ns/op\t 4964415 B/op\t      22 allocs/op",
            "extra": "229 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_TimeRange - ns/op",
            "value": 5241956,
            "unit": "ns/op",
            "extra": "229 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_TimeRange - B/op",
            "value": 4964415,
            "unit": "B/op",
            "extra": "229 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_TimeRange - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "229 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/DenialRate",
            "value": 857.3,
            "unit": "ns/op\t     418 B/op\t       5 allocs/op",
            "extra": "1355193 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/DenialRate - ns/op",
            "value": 857.3,
            "unit": "ns/op",
            "extra": "1355193 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/DenialRate - B/op",
            "value": 418,
            "unit": "B/op",
            "extra": "1355193 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/DenialRate - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1355193 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ErrorRate",
            "value": 831.5,
            "unit": "ns/op\t     418 B/op\t       5 allocs/op",
            "extra": "1355762 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ErrorRate - ns/op",
            "value": 831.5,
            "unit": "ns/op",
            "extra": "1355762 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ErrorRate - B/op",
            "value": 418,
            "unit": "B/op",
            "extra": "1355762 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ErrorRate - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1355762 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ActionVelocity",
            "value": 861.3,
            "unit": "ns/op\t     427 B/op\t       5 allocs/op",
            "extra": "1344242 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ActionVelocity - ns/op",
            "value": 861.3,
            "unit": "ns/op",
            "extra": "1344242 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ActionVelocity - B/op",
            "value": 427,
            "unit": "B/op",
            "extra": "1344242 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/ActionVelocity - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1344242 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/BudgetBreach",
            "value": 861.8,
            "unit": "ns/op\t     431 B/op\t       5 allocs/op",
            "extra": "1284308 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/BudgetBreach - ns/op",
            "value": 861.8,
            "unit": "ns/op",
            "extra": "1284308 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/BudgetBreach - B/op",
            "value": 431,
            "unit": "B/op",
            "extra": "1284308 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/BudgetBreach - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1284308 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/StuckAgent",
            "value": 823.3,
            "unit": "ns/op\t     420 B/op\t       5 allocs/op",
            "extra": "1319131 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/StuckAgent - ns/op",
            "value": 823.3,
            "unit": "ns/op",
            "extra": "1319131 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/StuckAgent - B/op",
            "value": 420,
            "unit": "B/op",
            "extra": "1319131 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureAlert_AllConditions/StuckAgent - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1319131 times\n4 procs"
          },
          {
            "name": "BenchmarkAlertLifecycle_ConfigResolve",
            "value": 117662,
            "unit": "ns/op\t     925 B/op\t      10 allocs/op",
            "extra": "14845 times\n4 procs"
          },
          {
            "name": "BenchmarkAlertLifecycle_ConfigResolve - ns/op",
            "value": 117662,
            "unit": "ns/op",
            "extra": "14845 times\n4 procs"
          },
          {
            "name": "BenchmarkAlertLifecycle_ConfigResolve - B/op",
            "value": 925,
            "unit": "B/op",
            "extra": "14845 times\n4 procs"
          },
          {
            "name": "BenchmarkAlertLifecycle_ConfigResolve - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "14845 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_JSON",
            "value": 2134020,
            "unit": "ns/op\t 2189818 B/op\t    2065 allocs/op",
            "extra": "556 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_JSON - ns/op",
            "value": 2134020,
            "unit": "ns/op",
            "extra": "556 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_JSON - B/op",
            "value": 2189818,
            "unit": "B/op",
            "extra": "556 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_JSON - allocs/op",
            "value": 2065,
            "unit": "allocs/op",
            "extra": "556 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_CSV",
            "value": 1508329,
            "unit": "ns/op\t 1825443 B/op\t    3053 allocs/op",
            "extra": "806 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_CSV - ns/op",
            "value": 1508329,
            "unit": "ns/op",
            "extra": "806 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_CSV - B/op",
            "value": 1825443,
            "unit": "B/op",
            "extra": "806 times\n4 procs"
          },
          {
            "name": "BenchmarkExportActions_CSV - allocs/op",
            "value": 3053,
            "unit": "allocs/op",
            "extra": "806 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantActionIsolation",
            "value": 230978,
            "unit": "ns/op\t  262158 B/op\t      13 allocs/op",
            "extra": "5330 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantActionIsolation - ns/op",
            "value": 230978,
            "unit": "ns/op",
            "extra": "5330 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantActionIsolation - B/op",
            "value": 262158,
            "unit": "B/op",
            "extra": "5330 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantActionIsolation - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "5330 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_HighThroughput",
            "value": 1370,
            "unit": "ns/op\t     695 B/op\t       7 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_HighThroughput - ns/op",
            "value": 1370,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_HighThroughput - B/op",
            "value": 695,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordAction_HighThroughput - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_PaginationWalk",
            "value": 125276811,
            "unit": "ns/op\t115288042 B/op\t     917 allocs/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_PaginationWalk - ns/op",
            "value": 125276811,
            "unit": "ns/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_PaginationWalk - B/op",
            "value": 115288042,
            "unit": "B/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkQueryActions_PaginationWalk - allocs/op",
            "value": 917,
            "unit": "allocs/op",
            "extra": "8 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost",
            "value": 827.4,
            "unit": "ns/op\t     477 B/op\t       5 allocs/op",
            "extra": "1304707 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - ns/op",
            "value": 827.4,
            "unit": "ns/op",
            "extra": "1304707 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - B/op",
            "value": 477,
            "unit": "B/op",
            "extra": "1304707 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1304707 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace",
            "value": 236.8,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "5035713 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace - ns/op",
            "value": 236.8,
            "unit": "ns/op",
            "extra": "5035713 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "5035713 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "5035713 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10",
            "value": 2139,
            "unit": "ns/op\t    5320 B/op\t       8 allocs/op",
            "extra": "517810 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - ns/op",
            "value": 2139,
            "unit": "ns/op",
            "extra": "517810 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - B/op",
            "value": 5320,
            "unit": "B/op",
            "extra": "517810 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=10 - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "517810 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100",
            "value": 24684,
            "unit": "ns/op\t   43336 B/op\t      11 allocs/op",
            "extra": "48753 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - ns/op",
            "value": 24684,
            "unit": "ns/op",
            "extra": "48753 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - B/op",
            "value": 43336,
            "unit": "B/op",
            "extra": "48753 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "48753 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000",
            "value": 321834,
            "unit": "ns/op\t  346447 B/op\t      14 allocs/op",
            "extra": "3655 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - ns/op",
            "value": 321834,
            "unit": "ns/op",
            "extra": "3655 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - B/op",
            "value": 346447,
            "unit": "B/op",
            "extra": "3655 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_Scaling/n=1000 - allocs/op",
            "value": 14,
            "unit": "allocs/op",
            "extra": "3655 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost",
            "value": 213.2,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "5598350 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - ns/op",
            "value": 213.2,
            "unit": "ns/op",
            "extra": "5598350 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "5598350 times\n4 procs"
          },
          {
            "name": "BenchmarkDeregisterHost - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "5598350 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_LargeFleet",
            "value": 25426,
            "unit": "ns/op\t   17744 B/op\t      14 allocs/op",
            "extra": "47127 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_LargeFleet - ns/op",
            "value": 25426,
            "unit": "ns/op",
            "extra": "47127 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_LargeFleet - B/op",
            "value": 17744,
            "unit": "B/op",
            "extra": "47127 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_LargeFleet - allocs/op",
            "value": 14,
            "unit": "allocs/op",
            "extra": "47127 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_Parallel",
            "value": 3608,
            "unit": "ns/op\t    2384 B/op\t      11 allocs/op",
            "extra": "314922 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_Parallel - ns/op",
            "value": 3608,
            "unit": "ns/op",
            "extra": "314922 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_Parallel - B/op",
            "value": 2384,
            "unit": "B/op",
            "extra": "314922 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_Parallel - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "314922 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost_Parallel",
            "value": 1182,
            "unit": "ns/op\t     527 B/op\t       7 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost_Parallel - ns/op",
            "value": 1182,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost_Parallel - B/op",
            "value": 527,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterHost_Parallel - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHeartbeat_Parallel",
            "value": 186.1,
            "unit": "ns/op\t     160 B/op\t       1 allocs/op",
            "extra": "6443961 times\n4 procs"
          },
          {
            "name": "BenchmarkHeartbeat_Parallel - ns/op",
            "value": 186.1,
            "unit": "ns/op",
            "extra": "6443961 times\n4 procs"
          },
          {
            "name": "BenchmarkHeartbeat_Parallel - B/op",
            "value": 160,
            "unit": "B/op",
            "extra": "6443961 times\n4 procs"
          },
          {
            "name": "BenchmarkHeartbeat_Parallel - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6443961 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MinResources",
            "value": 236.2,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "5060076 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MinResources - ns/op",
            "value": 236.2,
            "unit": "ns/op",
            "extra": "5060076 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MinResources - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "5060076 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MinResources - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "5060076 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MaxResources",
            "value": 236.5,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "5088544 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MaxResources - ns/op",
            "value": 236.5,
            "unit": "ns/op",
            "extra": "5088544 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MaxResources - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "5088544 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_MaxResources - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "5088544 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/standard",
            "value": 3022,
            "unit": "ns/op\t    1232 B/op\t      10 allocs/op",
            "extra": "378090 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/standard - ns/op",
            "value": 3022,
            "unit": "ns/op",
            "extra": "378090 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/standard - B/op",
            "value": 1232,
            "unit": "B/op",
            "extra": "378090 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/standard - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "378090 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/hardened",
            "value": 2877,
            "unit": "ns/op\t    1232 B/op\t      10 allocs/op",
            "extra": "402096 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/hardened - ns/op",
            "value": 2877,
            "unit": "ns/op",
            "extra": "402096 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/hardened - B/op",
            "value": 1232,
            "unit": "B/op",
            "extra": "402096 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/hardened - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "402096 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/isolated",
            "value": 2350,
            "unit": "ns/op\t     464 B/op\t       8 allocs/op",
            "extra": "488230 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/isolated - ns/op",
            "value": 2350,
            "unit": "ns/op",
            "extra": "488230 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/isolated - B/op",
            "value": 464,
            "unit": "B/op",
            "extra": "488230 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_TierFiltered/isolated - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "488230 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_NearCapacity",
            "value": 12580,
            "unit": "ns/op\t     119 B/op\t       0 allocs/op",
            "extra": "91719 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_NearCapacity - ns/op",
            "value": 12580,
            "unit": "ns/op",
            "extra": "91719 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_NearCapacity - B/op",
            "value": 119,
            "unit": "B/op",
            "extra": "91719 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_NearCapacity - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "91719 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureWarmPool",
            "value": 85,
            "unit": "ns/op\t      96 B/op\t       2 allocs/op",
            "extra": "14001516 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureWarmPool - ns/op",
            "value": 85,
            "unit": "ns/op",
            "extra": "14001516 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureWarmPool - B/op",
            "value": 96,
            "unit": "B/op",
            "extra": "14001516 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureWarmPool - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "14001516 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_WarmPoolHit",
            "value": 1055,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_WarmPoolHit - ns/op",
            "value": 1055,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_WarmPoolHit - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_WarmPoolHit - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=100",
            "value": 25011,
            "unit": "ns/op\t   43336 B/op\t      11 allocs/op",
            "extra": "47562 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=100 - ns/op",
            "value": 25011,
            "unit": "ns/op",
            "extra": "47562 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=100 - B/op",
            "value": 43336,
            "unit": "B/op",
            "extra": "47562 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "47562 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=500",
            "value": 147450,
            "unit": "ns/op\t  190794 B/op\t      13 allocs/op",
            "extra": "8206 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=500 - ns/op",
            "value": 147450,
            "unit": "ns/op",
            "extra": "8206 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=500 - B/op",
            "value": 190794,
            "unit": "B/op",
            "extra": "8206 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=500 - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "8206 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=2000",
            "value": 879534,
            "unit": "ns/op\t  895327 B/op\t      16 allocs/op",
            "extra": "1483 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=2000 - ns/op",
            "value": 879534,
            "unit": "ns/op",
            "extra": "1483 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=2000 - B/op",
            "value": 895327,
            "unit": "B/op",
            "extra": "1483 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=2000 - allocs/op",
            "value": 16,
            "unit": "allocs/op",
            "extra": "1483 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=5000",
            "value": 3675318,
            "unit": "ns/op\t 3631494 B/op\t      22 allocs/op",
            "extra": "322 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=5000 - ns/op",
            "value": 3675318,
            "unit": "ns/op",
            "extra": "322 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=5000 - B/op",
            "value": 3631494,
            "unit": "B/op",
            "extra": "322 times\n4 procs"
          },
          {
            "name": "BenchmarkListHosts_LargeFleet/n=5000 - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "322 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=100",
            "value": 4102,
            "unit": "ns/op\t     728 B/op\t       9 allocs/op",
            "extra": "285147 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=100 - ns/op",
            "value": 4102,
            "unit": "ns/op",
            "extra": "285147 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=100 - B/op",
            "value": 728,
            "unit": "B/op",
            "extra": "285147 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=100 - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "285147 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=500",
            "value": 19562,
            "unit": "ns/op\t     728 B/op\t       9 allocs/op",
            "extra": "61646 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=500 - ns/op",
            "value": 19562,
            "unit": "ns/op",
            "extra": "61646 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=500 - B/op",
            "value": 728,
            "unit": "B/op",
            "extra": "61646 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=500 - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "61646 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=2000",
            "value": 88782,
            "unit": "ns/op\t     728 B/op\t       9 allocs/op",
            "extra": "13491 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=2000 - ns/op",
            "value": 88782,
            "unit": "ns/op",
            "extra": "13491 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=2000 - B/op",
            "value": 728,
            "unit": "B/op",
            "extra": "13491 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCapacity_LargeFleet/n=2000 - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "13491 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedHeartbeatAndPlacement",
            "value": 1613,
            "unit": "ns/op\t    1061 B/op\t       3 allocs/op",
            "extra": "705241 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedHeartbeatAndPlacement - ns/op",
            "value": 1613,
            "unit": "ns/op",
            "extra": "705241 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedHeartbeatAndPlacement - B/op",
            "value": 1061,
            "unit": "B/op",
            "extra": "705241 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedHeartbeatAndPlacement - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "705241 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_HighContention",
            "value": 868.4,
            "unit": "ns/op\t     464 B/op\t       8 allocs/op",
            "extra": "1397392 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_HighContention - ns/op",
            "value": 868.4,
            "unit": "ns/op",
            "extra": "1397392 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_HighContention - B/op",
            "value": 464,
            "unit": "B/op",
            "extra": "1397392 times\n4 procs"
          },
          {
            "name": "BenchmarkPlaceWorkspace_HighContention - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "1397392 times\n4 procs"
          },
          {
            "name": "BenchmarkHostLifecycle_RegisterDeregister",
            "value": 762.4,
            "unit": "ns/op\t     944 B/op\t      11 allocs/op",
            "extra": "1569883 times\n4 procs"
          },
          {
            "name": "BenchmarkHostLifecycle_RegisterDeregister - ns/op",
            "value": 762.4,
            "unit": "ns/op",
            "extra": "1569883 times\n4 procs"
          },
          {
            "name": "BenchmarkHostLifecycle_RegisterDeregister - B/op",
            "value": 944,
            "unit": "B/op",
            "extra": "1569883 times\n4 procs"
          },
          {
            "name": "BenchmarkHostLifecycle_RegisterDeregister - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "1569883 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage",
            "value": 455.9,
            "unit": "ns/op\t     384 B/op\t       3 allocs/op",
            "extra": "2700270 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - ns/op",
            "value": 455.9,
            "unit": "ns/op",
            "extra": "2700270 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - B/op",
            "value": 384,
            "unit": "B/op",
            "extra": "2700270 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2700270 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget",
            "value": 67.8,
            "unit": "ns/op\t     160 B/op\t       1 allocs/op",
            "extra": "18400141 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - ns/op",
            "value": 67.8,
            "unit": "ns/op",
            "extra": "18400141 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - B/op",
            "value": 160,
            "unit": "B/op",
            "extra": "18400141 times\n4 procs"
          },
          {
            "name": "BenchmarkGetBudget - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "18400141 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget",
            "value": 464.4,
            "unit": "ns/op\t     536 B/op\t       5 allocs/op",
            "extra": "2576144 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - ns/op",
            "value": 464.4,
            "unit": "ns/op",
            "extra": "2576144 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - B/op",
            "value": 536,
            "unit": "B/op",
            "extra": "2576144 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2576144 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget",
            "value": 94.32,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "12449899 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - ns/op",
            "value": 94.32,
            "unit": "ns/op",
            "extra": "12449899 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12449899 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "12449899 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate",
            "value": 426.3,
            "unit": "ns/op\t     385 B/op\t       3 allocs/op",
            "extra": "2631322 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - ns/op",
            "value": 426.3,
            "unit": "ns/op",
            "extra": "2631322 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - B/op",
            "value": 385,
            "unit": "B/op",
            "extra": "2631322 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2631322 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel",
            "value": 567.6,
            "unit": "ns/op\t     393 B/op\t       3 allocs/op",
            "extra": "2227893 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - ns/op",
            "value": 567.6,
            "unit": "ns/op",
            "extra": "2227893 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - B/op",
            "value": 393,
            "unit": "B/op",
            "extra": "2227893 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2227893 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel",
            "value": 80.99,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "14732767 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - ns/op",
            "value": 80.99,
            "unit": "ns/op",
            "extra": "14732767 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "14732767 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "14732767 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit",
            "value": 99.89,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11782612 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - ns/op",
            "value": 99.89,
            "unit": "ns/op",
            "extra": "11782612 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11782612 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NearLimit - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11782612 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt",
            "value": 100.3,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11778381 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - ns/op",
            "value": 100.3,
            "unit": "ns/op",
            "extra": "11778381 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11778381 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Halt - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11778381 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn",
            "value": 100.8,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "11920279 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - ns/op",
            "value": 100.8,
            "unit": "ns/op",
            "extra": "11920279 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "11920279 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_OverLimit_Warn - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "11920279 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget",
            "value": 35.68,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "33877194 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - ns/op",
            "value": 35.68,
            "unit": "ns/op",
            "extra": "33877194 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "33877194 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_NoBudget - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "33877194 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert",
            "value": 476.1,
            "unit": "ns/op\t     536 B/op\t       5 allocs/op",
            "extra": "2521748 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - ns/op",
            "value": 476.1,
            "unit": "ns/op",
            "extra": "2521748 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - B/op",
            "value": 536,
            "unit": "B/op",
            "extra": "2521748 times\n4 procs"
          },
          {
            "name": "BenchmarkSetBudget_Upsert - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2521748 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel",
            "value": 566.3,
            "unit": "ns/op\t     393 B/op\t       3 allocs/op",
            "extra": "2243127 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - ns/op",
            "value": 566.3,
            "unit": "ns/op",
            "extra": "2243127 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - B/op",
            "value": 393,
            "unit": "B/op",
            "extra": "2243127 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_WithBudgetUpdate_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2243127 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes",
            "value": 24940,
            "unit": "ns/op\t    1816 B/op\t      19 allocs/op",
            "extra": "47845 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - ns/op",
            "value": 24940,
            "unit": "ns/op",
            "extra": "47845 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - B/op",
            "value": 1816,
            "unit": "B/op",
            "extra": "47845 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_ManyResourceTypes - allocs/op",
            "value": 19,
            "unit": "allocs/op",
            "extra": "47845 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset",
            "value": 181847,
            "unit": "ns/op\t     368 B/op\t       7 allocs/op",
            "extra": "6603 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - ns/op",
            "value": 181847,
            "unit": "ns/op",
            "extra": "6603 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - B/op",
            "value": 368,
            "unit": "B/op",
            "extra": "6603 times\n4 procs"
          },
          {
            "name": "BenchmarkGetCostReport_LargeDataset - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "6603 times\n4 procs"
          },
          {
            "name": "BenchmarkBudgetLifecycle_SetCheckRecord",
            "value": 1105,
            "unit": "ns/op\t    1448 B/op\t      15 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkBudgetLifecycle_SetCheckRecord - ns/op",
            "value": 1105,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkBudgetLifecycle_SetCheckRecord - B/op",
            "value": 1448,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkBudgetLifecycle_SetCheckRecord - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation",
            "value": 89.22,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "14676252 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - ns/op",
            "value": 89.22,
            "unit": "ns/op",
            "extra": "14676252 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "14676252 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantBudgetIsolation - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "14676252 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency",
            "value": 657.7,
            "unit": "ns/op\t     392 B/op\t       4 allocs/op",
            "extra": "1735406 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - ns/op",
            "value": 657.7,
            "unit": "ns/op",
            "extra": "1735406 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - B/op",
            "value": 392,
            "unit": "B/op",
            "extra": "1735406 times\n4 procs"
          },
          {
            "name": "BenchmarkRecordUsage_HighFrequency - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1735406 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold",
            "value": 96.58,
            "unit": "ns/op\t     208 B/op\t       2 allocs/op",
            "extra": "12378451 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - ns/op",
            "value": 96.58,
            "unit": "ns/op",
            "extra": "12378451 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "12378451 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckBudget_WarningThreshold - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "12378451 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload",
            "value": 2733,
            "unit": "ns/op\t      32 B/op\t       2 allocs/op",
            "extra": "432367 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - ns/op",
            "value": 2733,
            "unit": "ns/op",
            "extra": "432367 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - B/op",
            "value": 32,
            "unit": "B/op",
            "extra": "432367 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_SmallPayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "432367 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload",
            "value": 7626308,
            "unit": "ns/op\t   49219 B/op\t       2 allocs/op",
            "extra": "157 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_LargePayload - ns/op",
            "value": 7626308,
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
            "value": 193291,
            "unit": "ns/op\t    1414 B/op\t       1 allocs/op",
            "extra": "6358 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - ns/op",
            "value": 193291,
            "unit": "ns/op",
            "extra": "6358 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - B/op",
            "value": 1414,
            "unit": "B/op",
            "extra": "6358 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoMatch - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "6358 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy",
            "value": 24.76,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "49421542 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - ns/op",
            "value": 24.76,
            "unit": "ns/op",
            "extra": "49421542 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "49421542 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "49421542 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel",
            "value": 4697,
            "unit": "ns/op\t     128 B/op\t       3 allocs/op",
            "extra": "257991 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - ns/op",
            "value": 4697,
            "unit": "ns/op",
            "extra": "257991 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - B/op",
            "value": 128,
            "unit": "B/op",
            "extra": "257991 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "257991 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B",
            "value": 5058,
            "unit": "ns/op\t      48 B/op\t       2 allocs/op",
            "extra": "245466 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - ns/op",
            "value": 5058,
            "unit": "ns/op",
            "extra": "245466 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "245466 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10B - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "245466 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B",
            "value": 16171,
            "unit": "ns/op\t     128 B/op\t       2 allocs/op",
            "extra": "73460 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - ns/op",
            "value": 16171,
            "unit": "ns/op",
            "extra": "73460 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - B/op",
            "value": 128,
            "unit": "B/op",
            "extra": "73460 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100B - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "73460 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB",
            "value": 151220,
            "unit": "ns/op\t    1045 B/op\t       2 allocs/op",
            "extra": "7284 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - ns/op",
            "value": 151220,
            "unit": "ns/op",
            "extra": "7284 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - B/op",
            "value": 1045,
            "unit": "B/op",
            "extra": "7284 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7284 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB",
            "value": 1549787,
            "unit": "ns/op\t   10420 B/op\t       2 allocs/op",
            "extra": "774 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - ns/op",
            "value": 1549787,
            "unit": "ns/op",
            "extra": "774 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/10KB - B/op",
            "value": 10420,
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
            "value": 16211005,
            "unit": "ns/op\t  106535 B/op\t       2 allocs/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - ns/op",
            "value": 16211005,
            "unit": "ns/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - B/op",
            "value": 106535,
            "unit": "B/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/100KB - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "72 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB",
            "value": 165440423,
            "unit": "ns/op\t 1049776 B/op\t       8 allocs/op",
            "extra": "7 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_PayloadScaling/1MB - ns/op",
            "value": 165440423,
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
            "name": "BenchmarkClassifyData_AllPatterns/email",
            "value": 139085,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8509 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - ns/op",
            "value": 139085,
            "unit": "ns/op",
            "extra": "8509 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8509 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/email - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8509 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone",
            "value": 138162,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8839 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - ns/op",
            "value": 138162,
            "unit": "ns/op",
            "extra": "8839 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8839 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/phone - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8839 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn",
            "value": 149016,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8116 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - ns/op",
            "value": 149016,
            "unit": "ns/op",
            "extra": "8116 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8116 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/ssn - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8116 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card",
            "value": 149000,
            "unit": "ns/op\t    1040 B/op\t       2 allocs/op",
            "extra": "8086 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - ns/op",
            "value": 149000,
            "unit": "ns/op",
            "extra": "8086 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - B/op",
            "value": 1040,
            "unit": "B/op",
            "extra": "8086 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/credit_card - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8086 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key",
            "value": 148589,
            "unit": "ns/op\t    1044 B/op\t       2 allocs/op",
            "extra": "8186 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - ns/op",
            "value": 148589,
            "unit": "ns/op",
            "extra": "8186 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - B/op",
            "value": 1044,
            "unit": "B/op",
            "extra": "8186 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_AllPatterns/aws_key - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8186 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns",
            "value": 12045,
            "unit": "ns/op\t     272 B/op\t       4 allocs/op",
            "extra": "100212 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - ns/op",
            "value": 12045,
            "unit": "ns/op",
            "extra": "100212 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - B/op",
            "value": 272,
            "unit": "B/op",
            "extra": "100212 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_MultiplePatterns - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "100212 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns",
            "value": 160620,
            "unit": "ns/op\t    1029 B/op\t       1 allocs/op",
            "extra": "7419 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - ns/op",
            "value": 160620,
            "unit": "ns/op",
            "extra": "7419 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - B/op",
            "value": 1029,
            "unit": "B/op",
            "extra": "7419 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_NoPatterns - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "7419 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns",
            "value": 231268,
            "unit": "ns/op\t   11016 B/op\t       4 allocs/op",
            "extra": "5095 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - ns/op",
            "value": 231268,
            "unit": "ns/op",
            "extra": "5095 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - B/op",
            "value": 11016,
            "unit": "B/op",
            "extra": "5095 times\n4 procs"
          },
          {
            "name": "BenchmarkClassifyData_DensePatterns - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "5095 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public",
            "value": 15.55,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "73394778 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - ns/op",
            "value": 15.55,
            "unit": "ns/op",
            "extra": "73394778 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "73394778 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Public - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "73394778 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal",
            "value": 15.89,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "73332048 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - ns/op",
            "value": 15.89,
            "unit": "ns/op",
            "extra": "73332048 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "73332048 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Internal - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "73332048 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential",
            "value": 24.13,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "49707073 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - ns/op",
            "value": 24.13,
            "unit": "ns/op",
            "extra": "49707073 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "49707073 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Confidential - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "49707073 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted",
            "value": 15.52,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "70837114 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - ns/op",
            "value": 15.52,
            "unit": "ns/op",
            "extra": "70837114 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "70837114 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_AllClassifications/Restricted - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "70837114 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api",
            "value": 25.16,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "46837947 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - ns/op",
            "value": 25.16,
            "unit": "ns/op",
            "extra": "46837947 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "46837947 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/internal-api - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "46837947 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage",
            "value": 26.3,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "45363956 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - ns/op",
            "value": 26.3,
            "unit": "ns/op",
            "extra": "45363956 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "45363956 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/secure-storage - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "45363956 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log",
            "value": 23.83,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "50896088 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - ns/op",
            "value": 23.83,
            "unit": "ns/op",
            "extra": "50896088 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "50896088 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_ApprovedDestinations/audit-log - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "50896088 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api",
            "value": 26.99,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "43615122 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - ns/op",
            "value": 26.99,
            "unit": "ns/op",
            "extra": "43615122 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "43615122 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/external-api - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "43615122 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket",
            "value": 27.63,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "44754958 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - ns/op",
            "value": 27.63,
            "unit": "ns/op",
            "extra": "44754958 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "44754958 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/public-bucket - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "44754958 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service",
            "value": 28.7,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "41171578 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - ns/op",
            "value": 28.7,
            "unit": "ns/op",
            "extra": "41171578 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "41171578 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_DeniedDestinations/unknown-service - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "41171578 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel",
            "value": 11.57,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCheckPolicy_Parallel - ns/op",
            "value": 11.57,
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
            "value": 2841,
            "unit": "ns/op\t      96 B/op\t       3 allocs/op",
            "extra": "425919 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - ns/op",
            "value": 2841,
            "unit": "ns/op",
            "extra": "425919 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - B/op",
            "value": 96,
            "unit": "B/op",
            "extra": "425919 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "425919 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload",
            "value": 16201831,
            "unit": "ns/op\t  106620 B/op\t       2 allocs/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - ns/op",
            "value": 16201831,
            "unit": "ns/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - B/op",
            "value": 106620,
            "unit": "B/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_LargePayload - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "74 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload",
            "value": 639835,
            "unit": "ns/op\t    4116 B/op\t       1 allocs/op",
            "extra": "1875 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - ns/op",
            "value": 639835,
            "unit": "ns/op",
            "extra": "1875 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - B/op",
            "value": 4116,
            "unit": "B/op",
            "extra": "1875 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_CleanPayload - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1875 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload",
            "value": 263188,
            "unit": "ns/op\t    2744 B/op\t       3 allocs/op",
            "extra": "4630 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - ns/op",
            "value": 263188,
            "unit": "ns/op",
            "extra": "4630 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - B/op",
            "value": 2744,
            "unit": "B/op",
            "extra": "4630 times\n4 procs"
          },
          {
            "name": "BenchmarkInspectEgress_DirtyPayload - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4630 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule",
            "value": 1361,
            "unit": "ns/op\t    1079 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule - ns/op",
            "value": 1361,
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
            "value": 141.5,
            "unit": "ns/op\t     336 B/op\t       2 allocs/op",
            "extra": "8359018 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - ns/op",
            "value": 141.5,
            "unit": "ns/op",
            "extra": "8359018 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - B/op",
            "value": 336,
            "unit": "B/op",
            "extra": "8359018 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8359018 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100",
            "value": 37212,
            "unit": "ns/op\t   84105 B/op\t     111 allocs/op",
            "extra": "32077 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - ns/op",
            "value": 37212,
            "unit": "ns/op",
            "extra": "32077 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - B/op",
            "value": 84105,
            "unit": "B/op",
            "extra": "32077 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=100 - allocs/op",
            "value": 111,
            "unit": "allocs/op",
            "extra": "32077 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000",
            "value": 570128,
            "unit": "ns/op\t 1020249 B/op\t    1015 allocs/op",
            "extra": "2200 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - ns/op",
            "value": 570128,
            "unit": "ns/op",
            "extra": "2200 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - B/op",
            "value": 1020249,
            "unit": "B/op",
            "extra": "2200 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=1000 - allocs/op",
            "value": 1015,
            "unit": "allocs/op",
            "extra": "2200 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000",
            "value": 14404325,
            "unit": "ns/op\t14993704 B/op\t   10026 allocs/op",
            "extra": "92 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - ns/op",
            "value": 14404325,
            "unit": "ns/op",
            "extra": "92 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - B/op",
            "value": 14993704,
            "unit": "B/op",
            "extra": "92 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_Scaling/n=10000 - allocs/op",
            "value": 10026,
            "unit": "allocs/op",
            "extra": "92 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy",
            "value": 32025,
            "unit": "ns/op\t   40837 B/op\t     109 allocs/op",
            "extra": "41002 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - ns/op",
            "value": 32025,
            "unit": "ns/op",
            "extra": "41002 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - B/op",
            "value": 40837,
            "unit": "B/op",
            "extra": "41002 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy - allocs/op",
            "value": 109,
            "unit": "allocs/op",
            "extra": "41002 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_Parallel",
            "value": 1305,
            "unit": "ns/op\t     825 B/op\t       8 allocs/op",
            "extra": "888822 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_Parallel - ns/op",
            "value": 1305,
            "unit": "ns/op",
            "extra": "888822 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_Parallel - B/op",
            "value": 825,
            "unit": "B/op",
            "extra": "888822 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_Parallel - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "888822 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule_Parallel",
            "value": 157.1,
            "unit": "ns/op\t     336 B/op\t       2 allocs/op",
            "extra": "7560756 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule_Parallel - ns/op",
            "value": 157.1,
            "unit": "ns/op",
            "extra": "7560756 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule_Parallel - B/op",
            "value": 336,
            "unit": "B/op",
            "extra": "7560756 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRule_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7560756 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=1",
            "value": 755.2,
            "unit": "ns/op\t     664 B/op\t       5 allocs/op",
            "extra": "1586430 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=1 - ns/op",
            "value": 755.2,
            "unit": "ns/op",
            "extra": "1586430 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=1 - B/op",
            "value": 664,
            "unit": "B/op",
            "extra": "1586430 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=1 - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1586430 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=10",
            "value": 6115,
            "unit": "ns/op\t    8607 B/op\t      27 allocs/op",
            "extra": "197322 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=10 - ns/op",
            "value": 6115,
            "unit": "ns/op",
            "extra": "197322 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=10 - B/op",
            "value": 8607,
            "unit": "B/op",
            "extra": "197322 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=10 - allocs/op",
            "value": 27,
            "unit": "allocs/op",
            "extra": "197322 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=50",
            "value": 29345,
            "unit": "ns/op\t   42264 B/op\t     109 allocs/op",
            "extra": "40806 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=50 - ns/op",
            "value": 29345,
            "unit": "ns/op",
            "extra": "40806 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=50 - B/op",
            "value": 42264,
            "unit": "B/op",
            "extra": "40806 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=50 - allocs/op",
            "value": 109,
            "unit": "allocs/op",
            "extra": "40806 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=100",
            "value": 58561,
            "unit": "ns/op\t   85137 B/op\t     210 allocs/op",
            "extra": "20595 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=100 - ns/op",
            "value": 58561,
            "unit": "ns/op",
            "extra": "20595 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=100 - B/op",
            "value": 85137,
            "unit": "B/op",
            "extra": "20595 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=100 - allocs/op",
            "value": 210,
            "unit": "allocs/op",
            "extra": "20595 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=500",
            "value": 298254,
            "unit": "ns/op\t  401654 B/op\t    1013 allocs/op",
            "extra": "4081 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=500 - ns/op",
            "value": 298254,
            "unit": "ns/op",
            "extra": "4081 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=500 - B/op",
            "value": 401654,
            "unit": "B/op",
            "extra": "4081 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_Scaling/rules=500 - allocs/op",
            "value": 1013,
            "unit": "allocs/op",
            "extra": "4081 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_ComplexScopes",
            "value": 74987,
            "unit": "ns/op\t   74135 B/op\t     209 allocs/op",
            "extra": "16065 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_ComplexScopes - ns/op",
            "value": 74987,
            "unit": "ns/op",
            "extra": "16065 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_ComplexScopes - B/op",
            "value": 74135,
            "unit": "B/op",
            "extra": "16065 times\n4 procs"
          },
          {
            "name": "BenchmarkCompilePolicy_ComplexScopes - allocs/op",
            "value": 209,
            "unit": "allocs/op",
            "extra": "16065 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_SingleRule",
            "value": 446.5,
            "unit": "ns/op\t     776 B/op\t       7 allocs/op",
            "extra": "2686672 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_SingleRule - ns/op",
            "value": 446.5,
            "unit": "ns/op",
            "extra": "2686672 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_SingleRule - B/op",
            "value": 776,
            "unit": "B/op",
            "extra": "2686672 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_SingleRule - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "2686672 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=10",
            "value": 4118,
            "unit": "ns/op\t   13152 B/op\t      40 allocs/op",
            "extra": "286717 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=10 - ns/op",
            "value": 4118,
            "unit": "ns/op",
            "extra": "286717 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=10 - B/op",
            "value": 13152,
            "unit": "B/op",
            "extra": "286717 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=10 - allocs/op",
            "value": 40,
            "unit": "allocs/op",
            "extra": "286717 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=50",
            "value": 17795,
            "unit": "ns/op\t   56033 B/op\t     162 allocs/op",
            "extra": "68125 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=50 - ns/op",
            "value": 17795,
            "unit": "ns/op",
            "extra": "68125 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=50 - B/op",
            "value": 56033,
            "unit": "B/op",
            "extra": "68125 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=50 - allocs/op",
            "value": 162,
            "unit": "allocs/op",
            "extra": "68125 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=100",
            "value": 36444,
            "unit": "ns/op\t  114594 B/op\t     313 allocs/op",
            "extra": "32979 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=100 - ns/op",
            "value": 36444,
            "unit": "ns/op",
            "extra": "32979 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=100 - B/op",
            "value": 114594,
            "unit": "B/op",
            "extra": "32979 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=100 - allocs/op",
            "value": 313,
            "unit": "allocs/op",
            "extra": "32979 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=500",
            "value": 194059,
            "unit": "ns/op\t  501167 B/op\t    1515 allocs/op",
            "extra": "6116 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=500 - ns/op",
            "value": 194059,
            "unit": "ns/op",
            "extra": "6116 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=500 - B/op",
            "value": 501167,
            "unit": "B/op",
            "extra": "6116 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ManyRules/rules=500 - allocs/op",
            "value": 1515,
            "unit": "allocs/op",
            "extra": "6116 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ParameterCheck",
            "value": 475.6,
            "unit": "ns/op\t     776 B/op\t       7 allocs/op",
            "extra": "2527231 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ParameterCheck - ns/op",
            "value": 475.6,
            "unit": "ns/op",
            "extra": "2527231 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ParameterCheck - B/op",
            "value": 776,
            "unit": "B/op",
            "extra": "2527231 times\n4 procs"
          },
          {
            "name": "BenchmarkSimulatePolicy_ParameterCheck - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "2527231 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/tool_filter",
            "value": 1132,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/tool_filter - ns/op",
            "value": 1132,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/tool_filter - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/tool_filter - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/parameter_check",
            "value": 1159,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/parameter_check - ns/op",
            "value": 1159,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/parameter_check - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/parameter_check - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/rate_limit",
            "value": 1184,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/rate_limit - ns/op",
            "value": 1184,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/rate_limit - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/rate_limit - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/budget_limit",
            "value": 1167,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/budget_limit - ns/op",
            "value": 1167,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/budget_limit - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllTypes/budget_limit - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/allow",
            "value": 1203,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/allow - ns/op",
            "value": 1203,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/allow - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/allow - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/deny",
            "value": 1155,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/deny - ns/op",
            "value": 1155,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/deny - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/deny - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/escalate",
            "value": 1190,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/escalate - ns/op",
            "value": 1190,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/escalate - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/escalate - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/log",
            "value": 1202,
            "unit": "ns/op\t     839 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/log - ns/op",
            "value": 1202,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/log - B/op",
            "value": 839,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRule_AllActions/log - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_WithTypeFilter",
            "value": 126425,
            "unit": "ns/op\t  173226 B/op\t     262 allocs/op",
            "extra": "9476 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_WithTypeFilter - ns/op",
            "value": 126425,
            "unit": "ns/op",
            "extra": "9476 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_WithTypeFilter - B/op",
            "value": 173226,
            "unit": "B/op",
            "extra": "9476 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_WithTypeFilter - allocs/op",
            "value": 262,
            "unit": "allocs/op",
            "extra": "9476 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_EnabledOnlyFilter",
            "value": 247691,
            "unit": "ns/op\t  349070 B/op\t     513 allocs/op",
            "extra": "4852 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_EnabledOnlyFilter - ns/op",
            "value": 247691,
            "unit": "ns/op",
            "extra": "4852 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_EnabledOnlyFilter - B/op",
            "value": 349070,
            "unit": "B/op",
            "extra": "4852 times\n4 procs"
          },
          {
            "name": "BenchmarkListRules_EnabledOnlyFilter - allocs/op",
            "value": 513,
            "unit": "allocs/op",
            "extra": "4852 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateGuardrailSet",
            "value": 1651,
            "unit": "ns/op\t    1407 B/op\t       9 allocs/op",
            "extra": "881498 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateGuardrailSet - ns/op",
            "value": 1651,
            "unit": "ns/op",
            "extra": "881498 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateGuardrailSet - B/op",
            "value": 1407,
            "unit": "B/op",
            "extra": "881498 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateGuardrailSet - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "881498 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRuleAccess",
            "value": 141.1,
            "unit": "ns/op\t     336 B/op\t       2 allocs/op",
            "extra": "8588052 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRuleAccess - ns/op",
            "value": 141.1,
            "unit": "ns/op",
            "extra": "8588052 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRuleAccess - B/op",
            "value": 336,
            "unit": "B/op",
            "extra": "8588052 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRuleAccess - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "8588052 times\n4 procs"
          },
          {
            "name": "BenchmarkUpdateRule_Parallel",
            "value": 484.8,
            "unit": "ns/op\t     624 B/op\t       3 allocs/op",
            "extra": "2512264 times\n4 procs"
          },
          {
            "name": "BenchmarkUpdateRule_Parallel - ns/op",
            "value": 484.8,
            "unit": "ns/op",
            "extra": "2512264 times\n4 procs"
          },
          {
            "name": "BenchmarkUpdateRule_Parallel - B/op",
            "value": 624,
            "unit": "B/op",
            "extra": "2512264 times\n4 procs"
          },
          {
            "name": "BenchmarkUpdateRule_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2512264 times\n4 procs"
          },
          {
            "name": "BenchmarkDeleteRule_Throughput",
            "value": 956.2,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "1253409 times\n4 procs"
          },
          {
            "name": "BenchmarkDeleteRule_Throughput - ns/op",
            "value": 956.2,
            "unit": "ns/op",
            "extra": "1253409 times\n4 procs"
          },
          {
            "name": "BenchmarkDeleteRule_Throughput - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "1253409 times\n4 procs"
          },
          {
            "name": "BenchmarkDeleteRule_Throughput - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1253409 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest",
            "value": 1513,
            "unit": "ns/op\t    1127 B/op\t      10 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest - ns/op",
            "value": 1513,
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
            "value": 1497,
            "unit": "ns/op\t     976 B/op\t       7 allocs/op",
            "extra": "752569 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - ns/op",
            "value": 1497,
            "unit": "ns/op",
            "extra": "752569 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - B/op",
            "value": 976,
            "unit": "B/op",
            "extra": "752569 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "752569 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100",
            "value": 31276,
            "unit": "ns/op\t   78921 B/op\t      11 allocs/op",
            "extra": "38374 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - ns/op",
            "value": 31276,
            "unit": "ns/op",
            "extra": "38374 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - B/op",
            "value": 78921,
            "unit": "B/op",
            "extra": "38374 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "38374 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000",
            "value": 500387,
            "unit": "ns/op\t  922711 B/op\t      15 allocs/op",
            "extra": "2407 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - ns/op",
            "value": 500387,
            "unit": "ns/op",
            "extra": "2407 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - B/op",
            "value": 922711,
            "unit": "B/op",
            "extra": "2407 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=1000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "2407 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000",
            "value": 13285910,
            "unit": "ns/op\t13341858 B/op\t      26 allocs/op",
            "extra": "90 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - ns/op",
            "value": 13285910,
            "unit": "ns/op",
            "extra": "90 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - B/op",
            "value": 13341858,
            "unit": "B/op",
            "extra": "90 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_Scaling/n=10000 - allocs/op",
            "value": 26,
            "unit": "allocs/op",
            "extra": "90 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_Parallel",
            "value": 1677,
            "unit": "ns/op\t    1099 B/op\t      11 allocs/op",
            "extra": "745495 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_Parallel - ns/op",
            "value": 1677,
            "unit": "ns/op",
            "extra": "745495 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_Parallel - B/op",
            "value": 1099,
            "unit": "B/op",
            "extra": "745495 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_Parallel - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "745495 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRequest_Parallel",
            "value": 148.9,
            "unit": "ns/op\t     280 B/op\t       2 allocs/op",
            "extra": "7876056 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRequest_Parallel - ns/op",
            "value": 148.9,
            "unit": "ns/op",
            "extra": "7876056 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRequest_Parallel - B/op",
            "value": 280,
            "unit": "B/op",
            "extra": "7876056 times\n4 procs"
          },
          {
            "name": "BenchmarkGetRequest_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7876056 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest_Throughput",
            "value": 1634,
            "unit": "ns/op\t    1008 B/op\t       8 allocs/op",
            "extra": "754947 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest_Throughput - ns/op",
            "value": 1634,
            "unit": "ns/op",
            "extra": "754947 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest_Throughput - B/op",
            "value": 1008,
            "unit": "B/op",
            "extra": "754947 times\n4 procs"
          },
          {
            "name": "BenchmarkRespondToRequest_Throughput - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "754947 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/low",
            "value": 1446,
            "unit": "ns/op\t    1083 B/op\t      10 allocs/op",
            "extra": "858817 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/low - ns/op",
            "value": 1446,
            "unit": "ns/op",
            "extra": "858817 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/low - B/op",
            "value": 1083,
            "unit": "B/op",
            "extra": "858817 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/low - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "858817 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/normal",
            "value": 1564,
            "unit": "ns/op\t    1122 B/op\t      10 allocs/op",
            "extra": "942900 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/normal - ns/op",
            "value": 1564,
            "unit": "ns/op",
            "extra": "942900 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/normal - B/op",
            "value": 1122,
            "unit": "B/op",
            "extra": "942900 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/normal - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "942900 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/high",
            "value": 1604,
            "unit": "ns/op\t    1127 B/op\t      10 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/high - ns/op",
            "value": 1604,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/high - B/op",
            "value": 1127,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/high - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/critical",
            "value": 1446,
            "unit": "ns/op\t    1109 B/op\t      10 allocs/op",
            "extra": "920737 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/critical - ns/op",
            "value": 1446,
            "unit": "ns/op",
            "extra": "920737 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/critical - B/op",
            "value": 1109,
            "unit": "B/op",
            "extra": "920737 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllUrgencies/critical - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "920737 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/approval",
            "value": 1563,
            "unit": "ns/op\t    1119 B/op\t      10 allocs/op",
            "extra": "938155 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/approval - ns/op",
            "value": 1563,
            "unit": "ns/op",
            "extra": "938155 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/approval - B/op",
            "value": 1119,
            "unit": "B/op",
            "extra": "938155 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/approval - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "938155 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/question",
            "value": 1582,
            "unit": "ns/op\t    1127 B/op\t      10 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/question - ns/op",
            "value": 1582,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/question - B/op",
            "value": 1127,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/question - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/escalation",
            "value": 1531,
            "unit": "ns/op\t    1127 B/op\t      10 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/escalation - ns/op",
            "value": 1531,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/escalation - B/op",
            "value": 1127,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_AllTypes/escalation - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeOptions",
            "value": 1585,
            "unit": "ns/op\t    1346 B/op\t       9 allocs/op",
            "extra": "888796 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeOptions - ns/op",
            "value": 1585,
            "unit": "ns/op",
            "extra": "888796 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeOptions - B/op",
            "value": 1346,
            "unit": "B/op",
            "extra": "888796 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeOptions - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "888796 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeContext",
            "value": 1565,
            "unit": "ns/op\t    1127 B/op\t      10 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeContext - ns/op",
            "value": 1565,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeContext - B/op",
            "value": 1127,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateRequest_LargeContext - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_50K",
            "value": 61131382,
            "unit": "ns/op\t75633834 B/op\t      34 allocs/op",
            "extra": "19 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_50K - ns/op",
            "value": 61131382,
            "unit": "ns/op",
            "extra": "19 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_50K - B/op",
            "value": 75633834,
            "unit": "B/op",
            "extra": "19 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_50K - allocs/op",
            "value": 34,
            "unit": "allocs/op",
            "extra": "19 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithWorkspaceFilter",
            "value": 156298,
            "unit": "ns/op\t   78921 B/op\t      11 allocs/op",
            "extra": "6564 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithWorkspaceFilter - ns/op",
            "value": 156298,
            "unit": "ns/op",
            "extra": "6564 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithWorkspaceFilter - B/op",
            "value": 78921,
            "unit": "B/op",
            "extra": "6564 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithWorkspaceFilter - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "6564 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithStatusFilter",
            "value": 2453820,
            "unit": "ns/op\t 2102377 B/op\t      18 allocs/op",
            "extra": "463 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithStatusFilter - ns/op",
            "value": 2453820,
            "unit": "ns/op",
            "extra": "463 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithStatusFilter - B/op",
            "value": 2102377,
            "unit": "B/op",
            "extra": "463 times\n4 procs"
          },
          {
            "name": "BenchmarkListRequests_WithStatusFilter - allocs/op",
            "value": 18,
            "unit": "allocs/op",
            "extra": "463 times\n4 procs"
          },
          {
            "name": "BenchmarkRequestLifecycle_CreateRespondGet",
            "value": 1723,
            "unit": "ns/op\t    2784 B/op\t      27 allocs/op",
            "extra": "678622 times\n4 procs"
          },
          {
            "name": "BenchmarkRequestLifecycle_CreateRespondGet - ns/op",
            "value": 1723,
            "unit": "ns/op",
            "extra": "678622 times\n4 procs"
          },
          {
            "name": "BenchmarkRequestLifecycle_CreateRespondGet - B/op",
            "value": 2784,
            "unit": "B/op",
            "extra": "678622 times\n4 procs"
          },
          {
            "name": "BenchmarkRequestLifecycle_CreateRespondGet - allocs/op",
            "value": 27,
            "unit": "allocs/op",
            "extra": "678622 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureDeliveryChannel",
            "value": 169.9,
            "unit": "ns/op\t     160 B/op\t       2 allocs/op",
            "extra": "6996831 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureDeliveryChannel - ns/op",
            "value": 169.9,
            "unit": "ns/op",
            "extra": "6996831 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureDeliveryChannel - B/op",
            "value": 160,
            "unit": "B/op",
            "extra": "6996831 times\n4 procs"
          },
          {
            "name": "BenchmarkConfigureDeliveryChannel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "6996831 times\n4 procs"
          },
          {
            "name": "BenchmarkSetTimeoutPolicy",
            "value": 210.9,
            "unit": "ns/op\t     208 B/op\t       3 allocs/op",
            "extra": "5694429 times\n4 procs"
          },
          {
            "name": "BenchmarkSetTimeoutPolicy - ns/op",
            "value": 210.9,
            "unit": "ns/op",
            "extra": "5694429 times\n4 procs"
          },
          {
            "name": "BenchmarkSetTimeoutPolicy - B/op",
            "value": 208,
            "unit": "B/op",
            "extra": "5694429 times\n4 procs"
          },
          {
            "name": "BenchmarkSetTimeoutPolicy - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "5694429 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRequestIsolation",
            "value": 171.4,
            "unit": "ns/op\t     280 B/op\t       2 allocs/op",
            "extra": "7233908 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRequestIsolation - ns/op",
            "value": 171.4,
            "unit": "ns/op",
            "extra": "7233908 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRequestIsolation - B/op",
            "value": 280,
            "unit": "B/op",
            "extra": "7233908 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantRequestIsolation - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "7233908 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedRequestWorkload",
            "value": 727.5,
            "unit": "ns/op\t     485 B/op\t       5 allocs/op",
            "extra": "1684450 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedRequestWorkload - ns/op",
            "value": 727.5,
            "unit": "ns/op",
            "extra": "1684450 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedRequestWorkload - B/op",
            "value": 485,
            "unit": "B/op",
            "extra": "1684450 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedRequestWorkload - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1684450 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken",
            "value": 460.7,
            "unit": "ns/op\t     320 B/op\t       5 allocs/op",
            "extra": "2677094 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - ns/op",
            "value": 460.7,
            "unit": "ns/op",
            "extra": "2677094 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "2677094 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "2677094 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken",
            "value": 242.4,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "4798976 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - ns/op",
            "value": 242.4,
            "unit": "ns/op",
            "extra": "4798976 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "4798976 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4798976 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent",
            "value": 984.2,
            "unit": "ns/op\t     612 B/op\t       5 allocs/op",
            "extra": "1210212 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - ns/op",
            "value": 984.2,
            "unit": "ns/op",
            "extra": "1210212 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - B/op",
            "value": 612,
            "unit": "B/op",
            "extra": "1210212 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1210212 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent",
            "value": 125.4,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "9432463 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - ns/op",
            "value": 125.4,
            "unit": "ns/op",
            "extra": "9432463 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "9432463 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "9432463 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential",
            "value": 1609,
            "unit": "ns/op\t    1034 B/op\t      12 allocs/op",
            "extra": "862448 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - ns/op",
            "value": 1609,
            "unit": "ns/op",
            "extra": "862448 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - B/op",
            "value": 1034,
            "unit": "B/op",
            "extra": "862448 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "862448 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100",
            "value": 27082,
            "unit": "ns/op\t   54664 B/op\t      11 allocs/op",
            "extra": "43928 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - ns/op",
            "value": 27082,
            "unit": "ns/op",
            "extra": "43928 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - B/op",
            "value": 54664,
            "unit": "B/op",
            "extra": "43928 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=100 - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "43928 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000",
            "value": 414212,
            "unit": "ns/op\t  693652 B/op\t      15 allocs/op",
            "extra": "2647 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - ns/op",
            "value": 414212,
            "unit": "ns/op",
            "extra": "2647 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - B/op",
            "value": 693652,
            "unit": "B/op",
            "extra": "2647 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=1000 - allocs/op",
            "value": 15,
            "unit": "allocs/op",
            "extra": "2647 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000",
            "value": 10706842,
            "unit": "ns/op\t10556887 B/op\t      25 allocs/op",
            "extra": "111 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - ns/op",
            "value": 10706842,
            "unit": "ns/op",
            "extra": "111 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - B/op",
            "value": 10556887,
            "unit": "B/op",
            "extra": "111 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_Scaling/n=10000 - allocs/op",
            "value": 25,
            "unit": "allocs/op",
            "extra": "111 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent",
            "value": 848.7,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "1415738 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - ns/op",
            "value": 848.7,
            "unit": "ns/op",
            "extra": "1415738 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "1415738 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "1415738 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel",
            "value": 1625,
            "unit": "ns/op\t     988 B/op\t      11 allocs/op",
            "extra": "734271 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - ns/op",
            "value": 1625,
            "unit": "ns/op",
            "extra": "734271 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - B/op",
            "value": 988,
            "unit": "B/op",
            "extra": "734271 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_Parallel - allocs/op",
            "value": 11,
            "unit": "allocs/op",
            "extra": "734271 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel",
            "value": 242.2,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "4876940 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - ns/op",
            "value": 242.2,
            "unit": "ns/op",
            "extra": "4876940 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "4876940 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_Parallel - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4876940 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel",
            "value": 2039,
            "unit": "ns/op\t    1065 B/op\t      12 allocs/op",
            "extra": "575106 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - ns/op",
            "value": 2039,
            "unit": "ns/op",
            "extra": "575106 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - B/op",
            "value": 1065,
            "unit": "B/op",
            "extra": "575106 times\n4 procs"
          },
          {
            "name": "BenchmarkMintCredential_Parallel - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "575106 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite",
            "value": 560.3,
            "unit": "ns/op\t     382 B/op\t       3 allocs/op",
            "extra": "2326378 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - ns/op",
            "value": 560.3,
            "unit": "ns/op",
            "extra": "2326378 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - B/op",
            "value": 382,
            "unit": "B/op",
            "extra": "2326378 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedReadWrite - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2326378 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels",
            "value": 1030,
            "unit": "ns/op\t     631 B/op\t       5 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxLabels - ns/op",
            "value": 1030,
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
            "value": 1956,
            "unit": "ns/op\t    2457 B/op\t       7 allocs/op",
            "extra": "571681 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - ns/op",
            "value": 1956,
            "unit": "ns/op",
            "extra": "571681 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - B/op",
            "value": 2457,
            "unit": "B/op",
            "extra": "571681 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_MaxCapabilities - allocs/op",
            "value": 7,
            "unit": "allocs/op",
            "extra": "571681 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings",
            "value": 1124,
            "unit": "ns/op\t     679 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkRegisterAgent_LongStrings - ns/op",
            "value": 1124,
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
            "value": 244.3,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "5017921 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - ns/op",
            "value": 244.3,
            "unit": "ns/op",
            "extra": "5017921 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "5017921 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantConcurrentReads - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "5017921 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation",
            "value": 210069,
            "unit": "ns/op\t  226712 B/op\t      13 allocs/op",
            "extra": "5260 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - ns/op",
            "value": 210069,
            "unit": "ns/op",
            "extra": "5260 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - B/op",
            "value": 226712,
            "unit": "B/op",
            "extra": "5260 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantListIsolation - allocs/op",
            "value": 13,
            "unit": "allocs/op",
            "extra": "5260 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle",
            "value": 5486,
            "unit": "ns/op\t    2992 B/op\t      31 allocs/op",
            "extra": "192652 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - ns/op",
            "value": 5486,
            "unit": "ns/op",
            "extra": "192652 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - B/op",
            "value": 2992,
            "unit": "B/op",
            "extra": "192652 times\n4 procs"
          },
          {
            "name": "BenchmarkAgentLifecycle_FullCycle - allocs/op",
            "value": 31,
            "unit": "allocs/op",
            "extra": "192652 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn",
            "value": 1849,
            "unit": "ns/op\t    1127 B/op\t      14 allocs/op",
            "extra": "705157 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - ns/op",
            "value": 1849,
            "unit": "ns/op",
            "extra": "705157 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - B/op",
            "value": 1127,
            "unit": "B/op",
            "extra": "705157 times\n4 procs"
          },
          {
            "name": "BenchmarkCredentialChurn - allocs/op",
            "value": 14,
            "unit": "allocs/op",
            "extra": "705157 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel",
            "value": 212.5,
            "unit": "ns/op\t     320 B/op\t       5 allocs/op",
            "extra": "5639996 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - ns/op",
            "value": 212.5,
            "unit": "ns/op",
            "extra": "5639996 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - B/op",
            "value": 320,
            "unit": "B/op",
            "extra": "5639996 times\n4 procs"
          },
          {
            "name": "BenchmarkGenerateToken_Parallel - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "5639996 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel",
            "value": 113.3,
            "unit": "ns/op\t     192 B/op\t       3 allocs/op",
            "extra": "10369923 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - ns/op",
            "value": 113.3,
            "unit": "ns/op",
            "extra": "10369923 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - B/op",
            "value": 192,
            "unit": "B/op",
            "extra": "10369923 times\n4 procs"
          },
          {
            "name": "BenchmarkHashToken_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "10369923 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess",
            "value": 1729,
            "unit": "ns/op\t      48 B/op\t       1 allocs/op",
            "extra": "667526 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - ns/op",
            "value": 1729,
            "unit": "ns/op",
            "extra": "667526 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - B/op",
            "value": 48,
            "unit": "B/op",
            "extra": "667526 times\n4 procs"
          },
          {
            "name": "BenchmarkDeactivateAgent_ConcurrentCredentialAccess - allocs/op",
            "value": 1,
            "unit": "allocs/op",
            "extra": "667526 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K",
            "value": 120782034,
            "unit": "ns/op\t121304546 B/op\t      36 allocs/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - ns/op",
            "value": 120782034,
            "unit": "ns/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - B/op",
            "value": 121304546,
            "unit": "B/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkListAgents_100K - allocs/op",
            "value": 36,
            "unit": "allocs/op",
            "extra": "9 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants",
            "value": 263.7,
            "unit": "ns/op\t     256 B/op\t       2 allocs/op",
            "extra": "4564321 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - ns/op",
            "value": 263.7,
            "unit": "ns/op",
            "extra": "4564321 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - B/op",
            "value": 256,
            "unit": "B/op",
            "extra": "4564321 times\n4 procs"
          },
          {
            "name": "BenchmarkGetAgent_HighCardinalityTenants - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4564321 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_Parallel",
            "value": 1696,
            "unit": "ns/op\t    1280 B/op\t      10 allocs/op",
            "extra": "992053 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_Parallel - ns/op",
            "value": 1696,
            "unit": "ns/op",
            "extra": "992053 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_Parallel - B/op",
            "value": 1280,
            "unit": "B/op",
            "extra": "992053 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_Parallel - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "992053 times\n4 procs"
          },
          {
            "name": "BenchmarkGetTask_Parallel",
            "value": 254.7,
            "unit": "ns/op\t     544 B/op\t       3 allocs/op",
            "extra": "4705954 times\n4 procs"
          },
          {
            "name": "BenchmarkGetTask_Parallel - ns/op",
            "value": 254.7,
            "unit": "ns/op",
            "extra": "4705954 times\n4 procs"
          },
          {
            "name": "BenchmarkGetTask_Parallel - B/op",
            "value": 544,
            "unit": "B/op",
            "extra": "4705954 times\n4 procs"
          },
          {
            "name": "BenchmarkGetTask_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4705954 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_WithFullConfig",
            "value": 1805,
            "unit": "ns/op\t    1699 B/op\t       8 allocs/op",
            "extra": "736081 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_WithFullConfig - ns/op",
            "value": 1805,
            "unit": "ns/op",
            "extra": "736081 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_WithFullConfig - B/op",
            "value": 1699,
            "unit": "B/op",
            "extra": "736081 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_WithFullConfig - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "736081 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeInput",
            "value": 4525,
            "unit": "ns/op\t    3544 B/op\t      10 allocs/op",
            "extra": "268334 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeInput - ns/op",
            "value": 4525,
            "unit": "ns/op",
            "extra": "268334 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeInput - B/op",
            "value": 3544,
            "unit": "B/op",
            "extra": "268334 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeInput - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "268334 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeLabels",
            "value": 4167,
            "unit": "ns/op\t    3528 B/op\t      10 allocs/op",
            "extra": "315571 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeLabels - ns/op",
            "value": 4167,
            "unit": "ns/op",
            "extra": "315571 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeLabels - B/op",
            "value": 3528,
            "unit": "B/op",
            "extra": "315571 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateTask_LargeLabels - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "315571 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=100",
            "value": 51425,
            "unit": "ns/op\t  140234 B/op\t     211 allocs/op",
            "extra": "23197 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=100 - ns/op",
            "value": 51425,
            "unit": "ns/op",
            "extra": "23197 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=100 - B/op",
            "value": 140234,
            "unit": "B/op",
            "extra": "23197 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=100 - allocs/op",
            "value": 211,
            "unit": "allocs/op",
            "extra": "23197 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=1000",
            "value": 1096027,
            "unit": "ns/op\t 1619318 B/op\t    2016 allocs/op",
            "extra": "1128 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=1000 - ns/op",
            "value": 1096027,
            "unit": "ns/op",
            "extra": "1128 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=1000 - B/op",
            "value": 1619318,
            "unit": "B/op",
            "extra": "1128 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=1000 - allocs/op",
            "value": 2016,
            "unit": "allocs/op",
            "extra": "1128 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=10000",
            "value": 16887136,
            "unit": "ns/op\t23241897 B/op\t   20026 allocs/op",
            "extra": "80 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=10000 - ns/op",
            "value": 16887136,
            "unit": "ns/op",
            "extra": "80 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=10000 - B/op",
            "value": 23241897,
            "unit": "B/op",
            "extra": "80 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=10000 - allocs/op",
            "value": 20026,
            "unit": "allocs/op",
            "extra": "80 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=50000",
            "value": 85599478,
            "unit": "ns/op\t131398824 B/op\t  100034 allocs/op",
            "extra": "13 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=50000 - ns/op",
            "value": 85599478,
            "unit": "ns/op",
            "extra": "13 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=50000 - B/op",
            "value": 131398824,
            "unit": "B/op",
            "extra": "13 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_Scaling/n=50000 - allocs/op",
            "value": 100034,
            "unit": "allocs/op",
            "extra": "13 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithAgentFilter",
            "value": 89670,
            "unit": "ns/op\t  140234 B/op\t     211 allocs/op",
            "extra": "13347 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithAgentFilter - ns/op",
            "value": 89670,
            "unit": "ns/op",
            "extra": "13347 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithAgentFilter - B/op",
            "value": 140234,
            "unit": "B/op",
            "extra": "13347 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithAgentFilter - allocs/op",
            "value": 211,
            "unit": "allocs/op",
            "extra": "13347 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithStatusFilter",
            "value": 316654,
            "unit": "ns/op\t  547731 B/op\t     681 allocs/op",
            "extra": "3813 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithStatusFilter - ns/op",
            "value": 316654,
            "unit": "ns/op",
            "extra": "3813 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithStatusFilter - B/op",
            "value": 547731,
            "unit": "B/op",
            "extra": "3813 times\n4 procs"
          },
          {
            "name": "BenchmarkListTasks_WithStatusFilter - allocs/op",
            "value": 681,
            "unit": "allocs/op",
            "extra": "3813 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Running",
            "value": 1381,
            "unit": "ns/op\t    1088 B/op\t       6 allocs/op",
            "extra": "905064 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Running - ns/op",
            "value": 1381,
            "unit": "ns/op",
            "extra": "905064 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Running - B/op",
            "value": 1088,
            "unit": "B/op",
            "extra": "905064 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Running - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "905064 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Completed",
            "value": 1523,
            "unit": "ns/op\t    1136 B/op\t       8 allocs/op",
            "extra": "775957 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Completed - ns/op",
            "value": 1523,
            "unit": "ns/op",
            "extra": "775957 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Completed - B/op",
            "value": 1136,
            "unit": "B/op",
            "extra": "775957 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Completed - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "775957 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Failed",
            "value": 1495,
            "unit": "ns/op\t    1136 B/op\t       8 allocs/op",
            "extra": "800808 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Failed - ns/op",
            "value": 1495,
            "unit": "ns/op",
            "extra": "800808 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Failed - B/op",
            "value": 1136,
            "unit": "B/op",
            "extra": "800808 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Failed - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "800808 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_WaitingOnHuman",
            "value": 1219,
            "unit": "ns/op\t    1088 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_WaitingOnHuman - ns/op",
            "value": 1219,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_WaitingOnHuman - B/op",
            "value": 1088,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_WaitingOnHuman - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/WaitingOnHuman_Running",
            "value": 1173,
            "unit": "ns/op\t    1088 B/op\t       6 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/WaitingOnHuman_Running - ns/op",
            "value": 1173,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/WaitingOnHuman_Running - B/op",
            "value": 1088,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/WaitingOnHuman_Running - allocs/op",
            "value": 6,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Cancelled",
            "value": 1415,
            "unit": "ns/op\t     568 B/op\t       4 allocs/op",
            "extra": "842900 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Cancelled - ns/op",
            "value": 1415,
            "unit": "ns/op",
            "extra": "842900 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Cancelled - B/op",
            "value": 568,
            "unit": "B/op",
            "extra": "842900 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Pending_Cancelled - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "842900 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Cancelled",
            "value": 1168,
            "unit": "ns/op\t     568 B/op\t       4 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Cancelled - ns/op",
            "value": 1168,
            "unit": "ns/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Cancelled - B/op",
            "value": 568,
            "unit": "B/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStatusTransition_AllPaths/Running_Cancelled - allocs/op",
            "value": 4,
            "unit": "allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkTaskLifecycle_HappyPath",
            "value": 3828,
            "unit": "ns/op\t    3568 B/op\t      22 allocs/op",
            "extra": "291488 times\n4 procs"
          },
          {
            "name": "BenchmarkTaskLifecycle_HappyPath - ns/op",
            "value": 3828,
            "unit": "ns/op",
            "extra": "291488 times\n4 procs"
          },
          {
            "name": "BenchmarkTaskLifecycle_HappyPath - B/op",
            "value": 3568,
            "unit": "B/op",
            "extra": "291488 times\n4 procs"
          },
          {
            "name": "BenchmarkTaskLifecycle_HappyPath - allocs/op",
            "value": 22,
            "unit": "allocs/op",
            "extra": "291488 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantTaskIsolation",
            "value": 266.1,
            "unit": "ns/op\t     544 B/op\t       3 allocs/op",
            "extra": "4434050 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantTaskIsolation - ns/op",
            "value": 266.1,
            "unit": "ns/op",
            "extra": "4434050 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantTaskIsolation - B/op",
            "value": 544,
            "unit": "B/op",
            "extra": "4434050 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantTaskIsolation - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4434050 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedTaskWorkload",
            "value": 814.1,
            "unit": "ns/op\t     849 B/op\t       5 allocs/op",
            "extra": "1554936 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedTaskWorkload - ns/op",
            "value": 814.1,
            "unit": "ns/op",
            "extra": "1554936 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedTaskWorkload - B/op",
            "value": 849,
            "unit": "B/op",
            "extra": "1554936 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedTaskWorkload - allocs/op",
            "value": 5,
            "unit": "allocs/op",
            "extra": "1554936 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace",
            "value": 1313,
            "unit": "ns/op\t    1079 B/op\t       8 allocs/op",
            "extra": "1000000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace - ns/op",
            "value": 1313,
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
            "value": 187.8,
            "unit": "ns/op\t     456 B/op\t       3 allocs/op",
            "extra": "6281781 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - ns/op",
            "value": 187.8,
            "unit": "ns/op",
            "extra": "6281781 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - B/op",
            "value": 456,
            "unit": "B/op",
            "extra": "6281781 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "6281781 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100",
            "value": 41204,
            "unit": "ns/op\t  103177 B/op\t     111 allocs/op",
            "extra": "29101 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - ns/op",
            "value": 41204,
            "unit": "ns/op",
            "extra": "29101 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - B/op",
            "value": 103177,
            "unit": "B/op",
            "extra": "29101 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=100 - allocs/op",
            "value": 111,
            "unit": "allocs/op",
            "extra": "29101 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000",
            "value": 805500,
            "unit": "ns/op\t 1235942 B/op\t    1016 allocs/op",
            "extra": "1510 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - ns/op",
            "value": 805500,
            "unit": "ns/op",
            "extra": "1510 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - B/op",
            "value": 1235942,
            "unit": "B/op",
            "extra": "1510 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=1000 - allocs/op",
            "value": 1016,
            "unit": "allocs/op",
            "extra": "1510 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000",
            "value": 15495396,
            "unit": "ns/op\t18535342 B/op\t   10026 allocs/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - ns/op",
            "value": 15495396,
            "unit": "ns/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - B/op",
            "value": 18535342,
            "unit": "B/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_Scaling/n=10000 - allocs/op",
            "value": 10026,
            "unit": "allocs/op",
            "extra": "91 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace",
            "value": 284.1,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "4216765 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - ns/op",
            "value": 284.1,
            "unit": "ns/op",
            "extra": "4216765 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "4216765 times\n4 procs"
          },
          {
            "name": "BenchmarkTerminateWorkspace - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "4216765 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec",
            "value": 1555,
            "unit": "ns/op\t    1329 B/op\t       9 allocs/op",
            "extra": "757743 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - ns/op",
            "value": 1555,
            "unit": "ns/op",
            "extra": "757743 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - B/op",
            "value": 1329,
            "unit": "B/op",
            "extra": "757743 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_WithSpec - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "757743 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_Parallel",
            "value": 1707,
            "unit": "ns/op\t    1090 B/op\t      12 allocs/op",
            "extra": "751490 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_Parallel - ns/op",
            "value": 1707,
            "unit": "ns/op",
            "extra": "751490 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_Parallel - B/op",
            "value": 1090,
            "unit": "B/op",
            "extra": "751490 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_Parallel - allocs/op",
            "value": 12,
            "unit": "allocs/op",
            "extra": "751490 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace_Parallel",
            "value": 227.3,
            "unit": "ns/op\t     456 B/op\t       3 allocs/op",
            "extra": "5398000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace_Parallel - ns/op",
            "value": 227.3,
            "unit": "ns/op",
            "extra": "5398000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace_Parallel - B/op",
            "value": 456,
            "unit": "B/op",
            "extra": "5398000 times\n4 procs"
          },
          {
            "name": "BenchmarkGetWorkspace_Parallel - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "5398000 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_FullSpec",
            "value": 1727,
            "unit": "ns/op\t    1430 B/op\t      10 allocs/op",
            "extra": "715282 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_FullSpec - ns/op",
            "value": 1727,
            "unit": "ns/op",
            "extra": "715282 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_FullSpec - B/op",
            "value": 1430,
            "unit": "B/op",
            "extra": "715282 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_FullSpec - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "715282 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEnvVars",
            "value": 4363,
            "unit": "ns/op\t    3356 B/op\t      10 allocs/op",
            "extra": "303793 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEnvVars - ns/op",
            "value": 4363,
            "unit": "ns/op",
            "extra": "303793 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEnvVars - B/op",
            "value": 3356,
            "unit": "B/op",
            "extra": "303793 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEnvVars - allocs/op",
            "value": 10,
            "unit": "allocs/op",
            "extra": "303793 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEgressAllowlist",
            "value": 2189,
            "unit": "ns/op\t    2847 B/op\t       9 allocs/op",
            "extra": "636397 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEgressAllowlist - ns/op",
            "value": 2189,
            "unit": "ns/op",
            "extra": "636397 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEgressAllowlist - B/op",
            "value": 2847,
            "unit": "B/op",
            "extra": "636397 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeEgressAllowlist - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "636397 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeAllowedTools",
            "value": 1843,
            "unit": "ns/op\t    1944 B/op\t       9 allocs/op",
            "extra": "695448 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeAllowedTools - ns/op",
            "value": 1843,
            "unit": "ns/op",
            "extra": "695448 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeAllowedTools - B/op",
            "value": 1944,
            "unit": "B/op",
            "extra": "695448 times\n4 procs"
          },
          {
            "name": "BenchmarkCreateWorkspace_LargeAllowedTools - allocs/op",
            "value": 9,
            "unit": "allocs/op",
            "extra": "695448 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_50K",
            "value": 75110915,
            "unit": "ns/op\t105398180 B/op\t   50033 allocs/op",
            "extra": "15 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_50K - ns/op",
            "value": 75110915,
            "unit": "ns/op",
            "extra": "15 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_50K - B/op",
            "value": 105398180,
            "unit": "B/op",
            "extra": "15 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_50K - allocs/op",
            "value": 50033,
            "unit": "allocs/op",
            "extra": "15 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithAgentFilter",
            "value": 70410,
            "unit": "ns/op\t  103177 B/op\t     111 allocs/op",
            "extra": "17071 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithAgentFilter - ns/op",
            "value": 70410,
            "unit": "ns/op",
            "extra": "17071 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithAgentFilter - B/op",
            "value": 103177,
            "unit": "B/op",
            "extra": "17071 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithAgentFilter - allocs/op",
            "value": 111,
            "unit": "allocs/op",
            "extra": "17071 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithStatusFilter",
            "value": 429230,
            "unit": "ns/op\t  744758 B/op\t     680 allocs/op",
            "extra": "2646 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithStatusFilter - ns/op",
            "value": 429230,
            "unit": "ns/op",
            "extra": "2646 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithStatusFilter - B/op",
            "value": 744758,
            "unit": "B/op",
            "extra": "2646 times\n4 procs"
          },
          {
            "name": "BenchmarkListWorkspaces_WithStatusFilter - allocs/op",
            "value": 680,
            "unit": "allocs/op",
            "extra": "2646 times\n4 procs"
          },
          {
            "name": "BenchmarkWorkspaceLifecycle_CreateTerminate",
            "value": 2468,
            "unit": "ns/op\t    1168 B/op\t       8 allocs/op",
            "extra": "474026 times\n4 procs"
          },
          {
            "name": "BenchmarkWorkspaceLifecycle_CreateTerminate - ns/op",
            "value": 2468,
            "unit": "ns/op",
            "extra": "474026 times\n4 procs"
          },
          {
            "name": "BenchmarkWorkspaceLifecycle_CreateTerminate - B/op",
            "value": 1168,
            "unit": "B/op",
            "extra": "474026 times\n4 procs"
          },
          {
            "name": "BenchmarkWorkspaceLifecycle_CreateTerminate - allocs/op",
            "value": 8,
            "unit": "allocs/op",
            "extra": "474026 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantWorkspaceIsolation",
            "value": 244.7,
            "unit": "ns/op\t     456 B/op\t       3 allocs/op",
            "extra": "4949958 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantWorkspaceIsolation - ns/op",
            "value": 244.7,
            "unit": "ns/op",
            "extra": "4949958 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantWorkspaceIsolation - B/op",
            "value": 456,
            "unit": "B/op",
            "extra": "4949958 times\n4 procs"
          },
          {
            "name": "BenchmarkMultiTenantWorkspaceIsolation - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "4949958 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedWorkspaceWorkload",
            "value": 570.7,
            "unit": "ns/op\t     514 B/op\t       3 allocs/op",
            "extra": "2194512 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedWorkspaceWorkload - ns/op",
            "value": 570.7,
            "unit": "ns/op",
            "extra": "2194512 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedWorkspaceWorkload - B/op",
            "value": 514,
            "unit": "B/op",
            "extra": "2194512 times\n4 procs"
          },
          {
            "name": "BenchmarkMixedWorkspaceWorkload - allocs/op",
            "value": 3,
            "unit": "allocs/op",
            "extra": "2194512 times\n4 procs"
          }
        ]
      }
    ]
  }
}