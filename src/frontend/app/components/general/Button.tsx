import { twMerge } from "tailwind-merge";

const Button = ({
  children,
  disabled,
  className,
  onClick,
}: {
  children: React.ReactNode;
  disabled?: boolean;
  className?: string;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
}) => {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={twMerge(
        "bg-slate-600/80 hover:bg-slate-600 hover:shadow-sm hover:shadow-slate-500 transition-colors duration-100 ease-in px-4 rounded-xl py-2",
        className
      )}>
      {children}
    </button>
  );
};

export default Button;
