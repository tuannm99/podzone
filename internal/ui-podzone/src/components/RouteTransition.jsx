import React, { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import NProgress from 'nprogress';
import { Spin } from 'antd';

// Configure once
NProgress.configure({
  showSpinner: false,
  trickleSpeed: 150,
  minimum: 0.08,
});

export default function RouteTransition() {
  const location = useLocation();
  const first = useRef(true);
  const [overlay, setOverlay] = useState(false);

  useEffect(() => {
    // Skip first paint
    if (first.current) {
      first.current = false;
      return;
    }

    // Avoid flashing on ultra-fast transitions
    const startTimer = setTimeout(() => {
      NProgress.start();
      setOverlay(true);
    }, 120);

    // Ensure it's visible for a short minimum duration
    const doneTimer = setTimeout(() => {
      NProgress.done(true);
      setOverlay(false);
    }, 450);

    return () => {
      clearTimeout(startTimer);
      clearTimeout(doneTimer);
      NProgress.done(true);
      setOverlay(false);
    };
  }, [location.key]);

  // Optional overlay spinner (remove if you only want top bar)
  if (!overlay) return null;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(255,255,255,0.35)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 2000,
        pointerEvents: 'none',
      }}
    >
      <Spin />
    </div>
  );
}

