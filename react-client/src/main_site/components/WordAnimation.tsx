import { useState, useEffect } from "react";

interface WordCyclerProps {
  words: string[];
  eyebrow?: string;
  interval?: number;
}

const WordAnimation = ({ words, eyebrow, interval = 2200 }: WordCyclerProps) => {
  const [index, setIndex] = useState(0);
  const [visible, setVisible] = useState(true);

  useEffect(() => {
    const timer = setInterval(() => {
      setVisible(false);
      setTimeout(() => {
        setIndex((prev) => (prev + 1) % words.length);
        setVisible(true);
      }, 400);
    }, interval);
    return () => clearInterval(timer);
  }, [words, interval]);

  return (
    <div className="flex flex-col items-start">
      {eyebrow && (
        <p className="text-neutral-500 text-sm uppercase tracking-widest mb-6">
          {eyebrow}
        </p>
      )}
      <h1
        className="text-white font-medium"
        style={{
          fontSize: "3.5rem",
          lineHeight: 1,
          transition: "opacity 0.4s ease, transform 0.4s ease",
          opacity: visible ? 1 : 0,
          transform: visible ? "translateY(0px)" : "translateY(12px)",
        }}
      >
        {words[index]}
      </h1>
      <div
        className="bg-orange-500 mt-3"
        style={{
          height: "3px",
          borderRadius: "2px",
          transition: "width 0.4s ease",
          width: visible ? "75%" : "0%",
        }}
      />
    </div>
  );
};

export default WordAnimation;