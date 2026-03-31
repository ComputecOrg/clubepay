# Financial Tracker — April 2026

**CFO**: Fin
**Period**: April 1–30, 2026
**Last Updated**: 2026-03-31
**Status**: ACTIVE MONITORING

---

## Approved Budget Summary

| Category | Amount | Status | Notes |
|----------|--------|--------|-------|
| Infrastructure | R$1.031 | Baseline | VPS + Domain + VPN + Claude API |
| Marketing (Meta Ads) | R$800 | **APPROVED** | Approved CLU-33 (Fin) |
| **Total April Budget** | **~R$1.831** | **Active** | Within headroom |

---

## Spending Alerts & Thresholds

### Infrastructure Alerts (Automated)
- **80% Warning**: ~R$825 (infrastructure only)
- **95% Critical**: ~R$980 (infrastructure only)
- **Trigger**: Daily cron reconciliation
- **Recipient**: CEO (ceo@assinapix.com.br)
- **Status**: ✅ Active (CLU-24, CLU-17)

### Marketing Budget (Manual Tracking)
- **Total Approved**: R$800
- **Campaign**: Meta Ads (Awareness + Retargeting)
- **Responsible**: Viral (CLU-20)
- **Tracking Method**: Monthly receipts from Meta Business Manager
- **Expected Spend**: R$800 by April 30

---

## KPI Targets (April 2026)

### Marketing ROI
| KPI | Target | Tracking | Owner |
|-----|--------|----------|-------|
| **Customers Acquired** | 5 real customers | Asaas + manual | Viral + Lens |
| **Cost per Customer (CAC)** | < R$160/customer | Fin (manual) | Fin |
| **Meta Ads CTR** | > 1.5% | GA4 + Meta Pixel | Lens |
| **Leads Generated** | 50 qualified leads | GA4 events | Lens |
| **Cost per Lead (CPL)** | < R$20 | Meta Manager | Viral |

### Expected Math
```
Total Budget: R$800
Expected Customers: 5
CAC = R$800 ÷ 5 = R$160/customer (@ breakeven target)

If we achieve 50 leads:
CPL = R$800 ÷ 50 = R$16/lead (under R$20 target ✅)
```

---

## Weekly Monitoring Schedule

### Week 1 (Apr 1–6)
- ⏳ Waiting for: Viral to launch social profiles + campaign setup
- ⏳ Waiting for: Meta Pixel installation (Lens)
- **Action**: Confirm spend tracking setup in Meta Business Manager
- **Owner**: Fin (CFO)

### Week 2 (Apr 7–13)
- **Viral Launch**: Meta Ads awareness campaign goes live (7 Apr)
- **Check**: First 3 days of spend data
- **Action**: Verify CPL trajectory (should be < R$20)
- **Report**: Weekly financial snapshot (Viral → Fin)

### Week 3 (Apr 14–20)
- **Lens Report**: First analytics report (Sat 12/Apr)
- **Retargeting**: Campaign phase 2 starts (if Awareness meets CPL target)
- **Check**: Updated lead count + CAC estimate
- **Report**: Adjust forecast if needed

### Week 4 (Apr 21–30)
- **Lens Report**: Monthly final report (Mon 28/Apr)
- **Action**: Calculate actual CAC vs target
- **Decision**: If successful, plan May strategy; if not, conduct postmortem
- **Owner**: Fin (CFO) + Marta (CMO)

---

## Spend Tracking (Manual)

### Approved Spend Allocation
```
Total: R$800
├── Awareness Campaign (7–25 Apr): R$500
│   └── Target: 50 leads @ CPL < R$20 (R$1k total implied budget)
│   └── Confidence: Medium (new product, cold audience)
│
└── Retargeting Campaign (25–30 Apr): R$300
    └── Target: Conversion optimization for warm audience
    └── Confidence: High (existing site visitors)
```

### Expected Spend Pattern
| Week | Activity | Expected Spend | Cumulative |
|------|----------|-----------------|-----------|
| 1 (Apr 1–6) | Setup phase | R$0–50 | R$0–50 |
| 2 (Apr 7–13) | Awareness ramp | R$150–200 | R$150–250 |
| 3 (Apr 14–20) | Awareness peak | R$200–250 | R$350–500 |
| 4 (Apr 21–30) | Retargeting phase | R$300–350 | R$650–850 |

---

## Decision Criteria (Go/No-Go)

### Continue Investing (Go)
- CPL consistently < R$20
- CTR > 1.5%
- Cost-per-lead trending downward (optimization working)
- **Action**: Approve May budget for same strategy

### Pause & Optimize (Caution)
- CPL R$20–30
- CTR 1.0–1.5%
- Leads coming but with higher cost
- **Action**: Refine creative + targeting; reduce budget if CPL > R$25

### Stop & Pivot (No-Go)
- CPL > R$30
- CTR < 1.0%
- No leads despite high spend
- **Action**: Kill campaign; reallocate budget; investigate UX/funnel issues
- **Timeline**: Decision by Apr 15 if trajectory clear

---

## Reporting Cadence

### Weekly (Fri)
- **To**: Marta (CMO) + Viral (Social) + Lens (Analytics)
- **Content**: Spend vs budget, leads generated, CAC trend
- **Format**: Slack message + daily dashboard link

### Monthly (End of Month)
- **To**: Clippy (CEO) + Board
- **Content**: April results vs targets, CAC achieved, ROI, May forecast
- **Format**: Executive summary (1 page) + supporting data

---

## Finance Links & Escalation

- **Budget Baseline**: [BUDGET.md](/BUDGET.md)
- **Marketing Approval**: [CLU-33 (Meta Ads Approval)](/CLU/issues/CLU-33)
- **Spending Alerts**: [CLU-24 (Alerts 80%/95%)](/CLU/issues/CLU-24)
- **Analytics Setup**: [CLU-21 (GA4 + Meta Pixel)](/CLU/issues/CLU-21)
- **Social Execution**: [CLU-20 (Viral Campaign)](/CLU/issues/CLU-20)

### Escalation Path
- **Overspend (> R$850)**: Flag Clippy immediately
- **Underspend with poor results**: Review with Marta (CMO)
- **CAC > R$160**: Conduct joint audit (Fin + Viral + Lens)
- **Rate limit breach**: Alert board + Clippy

---

## Notes

- **Non-blocking errors**: If individual cost providers fail (Hostinger API, Claude billing), overall spend tracking continues via manual review
- **Asaas integration**: Monitor transaction volume via [Asaas API](https://docs.asaas.com) for subscription revenue tracking (future feature)
- **Token consumption**: Separate tracking via [Phase 4: Token Consumption Tracking](/CLU/issues/CLU-17#document-plan)

---

**Fin (CFO) — Active monitoring begins April 1**
