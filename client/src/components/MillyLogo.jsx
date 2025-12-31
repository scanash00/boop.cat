// Copyright 2025 boop.cat
// Licensed under the Apache License, Version 2.0
// See LICENSE file for details.

import React from 'react';
export default function MillyLogo({ size = 24, style = {}, ...props }) {
  return (
    <img
      src="/milly.png"
      alt="Milly mascot"
      width={size}
      height={size}
      style={{ imageRendering: 'pixelated', display: 'inline-block', ...style }}
      onError={(e) => {
        e.target.style.display = 'none';
      }}
      {...props}
    />
  );
}
