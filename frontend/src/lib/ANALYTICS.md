# Analytics Integration (GA4 + Meta Pixel)

Frontend analytics tracking for ClubePay using Google Analytics 4 (GA4) and Meta Pixel.

## Setup

### 1. Environment Variables

Add to `.env.local` (development) or deployment config:

```env
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
NEXT_PUBLIC_PIXEL_ID=123456789012345
```

### 2. Wrap App with Provider

In `src/app/layout.tsx`:

```tsx
import { AnalyticsProvider } from '@/lib/analyticsContext';

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="pt-BR" className="h-full">
      <AnalyticsProvider>
        <body className="min-h-full flex flex-col">{children}</body>
      </AnalyticsProvider>
    </html>
  );
}
```

## Usage

Use the `useAnalytics()` hook in any client component:

```tsx
'use client';

import { useAnalytics } from '@/hooks/useAnalytics';

export default function RegisterForm() {
  const analytics = useAnalytics();

  const handleSubmit = async (data: FormData) => {
    // Track user registration
    analytics.trackBusinessCreated(
      userId,
      businessName,
      industry
    );

    // ... submit form
  };

  return <form onSubmit={handleSubmit}>...</form>;
}
```

## Events

### GA4 Events

| Event | Method | Parameters |
|-------|--------|------------|
| Business Created | `trackBusinessCreated(userId, businessName, industry)` | user_id, business_name, industry |
| Subscription Created | `trackSubscriptionCreated(userId, planId, planName, referralCode?)` | user_id, plan_id, plan_name, referral_code* |
| Payment Received | `trackPaymentReceived(userId, amountInCents, planType)` | user_id, value (in R$), currency, plan_type |
| Validation Success | `trackValidationSuccess(userId)` | user_id |

### Meta Pixel Events

| Event | Method | Parameters |
|-------|--------|------------|
| Purchase | `trackMetaPixelPurchase(amountInCents, currency, contentId)` | value, currency, content_name, content_id |
| Lead | `trackMetaPixelLead(contentName)` | content_name |
| ViewContent | `trackMetaPixelViewContent(contentId, contentName)` | content_id, content_name |

*Optional parameter

## Testing

Run tests with:

```bash
npm test
```

Tests cover:
- Analytics class initialization
- GA4 event tracking
- Meta Pixel event tracking
- React hook integration
- Component-level integration

## Best Practices

1. **Always wrap components with `useAnalytics()` check**
   ```tsx
   const analytics = useAnalytics();
   // Won't throw even if gtag/fbq unavailable
   analytics.trackBusinessCreated(...);
   ```

2. **Track key user journeys**
   - Sign-up/registration
   - Plan creation
   - Subscription (with referral info if applicable)
   - Payment success/failure
   - Feature usage

3. **Don't track PII**
   - Avoid sending email, phone, SSN, etc.
   - Track only IDs, categories, and event metadata

4. **Use environment variables**
   - Never hardcode measurement IDs
   - Always check `NEXT_PUBLIC_GA_ID` and `NEXT_PUBLIC_PIXEL_ID`

## Viewing Analytics

- **GA4:** https://analytics.google.com
- **Meta Pixel:** https://business.facebook.com → Events Manager

## Troubleshooting

### Analytics not tracking

1. Check environment variables are set
2. Check browser console for errors
3. Verify gtag/fbq scripts loaded (Network tab)
4. Check GA4/Meta dashboard for test events

### Development Mode

In development, analytics initializes with no-op implementation if IDs aren't set. This prevents errors but events won't be tracked. Set IDs to test locally.

## References

- [GA4 Event Tracking](https://developers.google.com/analytics/devguides/collection/ga4/events)
- [Meta Pixel Events](https://developers.facebook.com/docs/facebook-pixel/reference)
