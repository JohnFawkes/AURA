export function UtilityBackground() {
  return (
    <div className='fixed inset-0 bg-gray-950 overflow-hidden -z-10'>
      <div className='absolute inset-0 bg-gradient-to-br from-background to-background' />
      <div className='absolute -inset-[10%]'>
        <div className='absolute top-[10%] left-[20%] w-[30%] aspect-square rounded-full bg-gradient-to-r from-purple-600/20 to-indigo-700/20 blur-3xl' />
        <div className='absolute top-[60%] left-[60%] w-[25%] aspect-square rounded-full bg-gradient-to-r from-amber-600/20 to-orange-700/20 blur-3xl' />
        <div className='absolute top-[40%] left-[30%] w-[20%] aspect-square rounded-full bg-gradient-to-r from-emerald-600/20 to-teal-700/20 blur-3xl' />
      </div>
      <div className='absolute inset-0 backdrop-blur-[5px]' />
    </div>
  );
}
