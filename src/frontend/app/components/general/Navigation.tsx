import { Link } from "@remix-run/react";

const navList = [
  {
    name: "Home",
    href: "/",
  },
  {
    name: "Albums",
    href: "/albums",
  },
  {
    name: "Songs",
    href: "/songs",
  },
  {
    name: "Manage",
    href: "/manage",
  },
];

const Navigation = () => {
  return (
    <>
      <header className="flex md:hidden w-screen h-20 p-4 bg-[#161617] items-center justify-between">
        <p className="text-slate-200 text-xl">D&apos;Finder</p>
      </header>
      <aside className="hidden md:flex flex-col h-screen w-64 py-4 px-2 bg-[#161617] items-center gap-4 min-h-fit">
        <p className="text-slate-200 text-xl">D&apos;Finder</p>
        <nav className="w-full">
          {navList.map((navItem) => (
            <Link key={navItem.href} to={navItem.href}>
              <button className="block w-full text-start p-2 hover:bg-white/10 rounded-md">{navItem.name}</button>
            </Link>
          ))}
        </nav>
      </aside>
    </>
  );
};

export default Navigation;
