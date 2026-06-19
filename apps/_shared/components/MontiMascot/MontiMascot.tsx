"use client";
import React from "react";

export type MontiMood = "happy" | "listening" | "thinking" | "speaking" | "concerned";

export interface MontiMascotProps {
  mood?: MontiMood;
  glow?: boolean;
  size?: number;
  bobbing?: boolean;
  className?: string;
}

/**
 * MontiMascot — SVG Shiba Inu mascot with optional 64px radial cyan glow
 * and idle bob animation (4s ease-in-out). DES-0008 §1.4, §1.5.
 */
export function MontiMascot({
  mood = "happy",
  glow = true,
  size = 240,
  bobbing = true,
  className = "",
}: MontiMascotProps) {
  return (
    <div
      className={`relative inline-flex items-center justify-center ${className}`}
      style={{ width: size, height: size }}
      data-mood={mood}
      data-testid="monti-mascot"
      aria-hidden="true"
    >
      {glow && (
        <div
          className="absolute inset-0 rounded-full"
          style={{
            background:
              "radial-gradient(circle, rgba(51,204,255,0.3) 0%, rgba(0,191,255,0.15) 40%, rgba(0,0,0,0) 70%)",
            filter: "blur(8px)",
            transform: "scale(1.35)",
          }}
        />
      )}
      <div className={bobbing ? "animate-monti-bob" : ""}>
        <svg width={size * 0.7} height={size * 0.7} viewBox="0 0 200 200" fill="none">
          {/* Head */}
          <ellipse cx="100" cy="110" rx="70" ry="62" fill="#E8A06A" />
          {/* Cheeks lighter */}
          <ellipse cx="100" cy="130" rx="50" ry="32" fill="#FFE5CC" />
          {/* Ears */}
          <path d="M40 70 L55 30 L75 60 Z" fill="#E8A06A" />
          <path d="M160 70 L145 30 L125 60 Z" fill="#E8A06A" />
          <path d="M50 60 L60 40 L70 58 Z" fill="#B8743C" />
          <path d="M150 60 L140 40 L130 58 Z" fill="#B8743C" />
          {/* Eyes */}
          <circle cx="75" cy="105" r="7" fill="#0B1438" />
          <circle cx="125" cy="105" r="7" fill="#0B1438" />
          <circle cx="73" cy="103" r="2" fill="#FFFFFF" />
          <circle cx="123" cy="103" r="2" fill="#FFFFFF" />
          {/* Nose */}
          <ellipse cx="100" cy="128" rx="8" ry="6" fill="#0B1438" />
          {/* Mouth — varies by mood */}
          {mood === "concerned" ? (
            <path d="M85 150 Q100 142 115 150" stroke="#0B1438" strokeWidth="3" fill="none" />
          ) : (
            <path d="M85 145 Q92 158 100 145 Q108 158 115 145" stroke="#0B1438" strokeWidth="3" fill="none" />
          )}
        </svg>
      </div>
    </div>
  );
}

export default MontiMascot;
