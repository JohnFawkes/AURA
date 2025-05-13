export function MeshBackground() {
  return (
    <div className='fixed inset-0 dark:bg-gray-950 overflow-hidden -z-10'>
      <div className='absolute -inset-[100%] opacity-50'>
        <div className='absolute top-1/3 left-1/4 w-[50%] aspect-square rounded-full bg-gradient-to-r from-dark-muted to-light-muted blur-3xl' />
        <div className='absolute top-2/3 left-1/2 w-[40%] aspect-square rounded-full bg-gradient-to-r from-ultra-dark-muted to-vibrant blur-3xl' />
        <div className='absolute top-1/4 right-1/4 w-[30%] aspect-square rounded-full bg-gradient-to-r from-transparent to-dark-vibrant blur-3xl' />
      </div>
    </div>
  );
}
