import React from 'react';

interface ConnectionIndicatorProps {
  connected: boolean;
}

export default function ConnectionIndicator({ connected }: ConnectionIndicatorProps) {
  return (
    <div className="connection-indicator">
      <span className={`connection-dot ${connected ? 'connected' : ''}`}></span>
      <span>{connected ? 'Connected' : 'Disconnected'}</span>
    </div>
  );
}
