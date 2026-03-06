interface SliderProps {
  label: string;
  value: number;
  min: number;
  max: number;
  step?: number;
  onChange: (value: number) => void;
  valueLabel?: string;
}

export const Slider = ({ label, value, min, max, step = 1, onChange, valueLabel }: SliderProps) => {
  const percent = ((value - min) / (max - min)) * 100;
  return (
    <div className="w-full">
      <div className="flex justify-between items-baseline mb-2">
        <label className="text-white text-sm font-regular">{label}</label>
        {valueLabel != null ? (
          <span className="text-neutral-400 text-sm">{valueLabel}</span>
        ) : (
          <span className="text-neutral-400 text-sm">{value}</span>
        )}
      </div>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full h-2 rounded-full appearance-none cursor-pointer bg-neutral-800 focus:outline-none focus:ring-0 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-4 [&::-webkit-slider-thumb]:h-4 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-orange-500 [&::-webkit-slider-thumb]:cursor-pointer [&::-webkit-slider-thumb]:border-2 [&::-webkit-slider-thumb]:border-orange-400 [&::-moz-range-thumb]:w-4 [&::-moz-range-thumb]:h-4 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-orange-500 [&::-moz-range-thumb]:border-2 [&::-moz-range-thumb]:border-orange-400 [&::-moz-range-thumb]:border-none"
        style={{
          background: `linear-gradient(to right, rgb(249 115 22) 0%, rgb(249 115 22) ${percent}%, rgb(38 38 38) ${percent}%, rgb(38 38 38) 100%)`,
        }}
      />
    </div>
  );
};
