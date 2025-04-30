export function AppBackground() {
  return (
    <div className='absolute inset-0 bg-gray-950 -z-20'>
      <div
        className='absolute inset-0 opacity-[0.07]'
        style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.65' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E")`,
          backgroundSize: "200px 200px",
        }}
      />
      <div className='absolute inset-0 bg-gradient-to-br from-blue-950/30 via-transparent to-purple-950/30' />
      <div className='absolute inset-0 bg-gradient-to-tr from-yellow-950/20 via-transparent to-pink-950/20' />
    </div>
  );
}
