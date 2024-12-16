import type { MetaFunction } from "@remix-run/node";

export const meta: MetaFunction = () => {
  return [{ title: "New Remix App" }, { name: "description", content: "Welcome to Remix!" }];
};

export default function Index() {
  return (
    <div className="flex h-screen w-0 grow items-center justify-center flex-col-reverse">
      <h1>Mencariiiii apapun dengan canggih :)</h1>
      <pre>{`
   .
  .
 . .
  ...
\\~~~~~/
 \\   /
  \\ /
   V
   |
   |
  ---`}</pre>
    </div>
  );
}
