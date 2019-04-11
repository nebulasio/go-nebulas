Function Flow
============
Function Flow is a library which targets on "ease parallel programming", especially expressing task parallelism. When we say "ease parallel programming", we are referring a lot of things, including

    1. easy-to-use programming interfaces;
    2. automatically performance tuning;
    3. tool chains with static program analysis.

In the part of easy used programming interfaces, we carefully desgined Function Flow to avoid any "anti-intuitive" programming statement. All things in expressing parallelism are done directly without anything like initialization, constraints of tasks, tricks of waiting expression and etc.

Besides, Function Flow also makes a huge effort to make parallel programming "safe" by leveraging compiling-time check, instead of runtime assert. This makes a successfully compiled Function Flow program satisfies some parallel constraints already. Such feature largely helps improving productivity. Also, we customize some compile errors to help users understanding why their programs are "invalid".

In the part of automatically performance tuning, Function Flow tries to get hints from static analysis to balance different choices in runtime subsystem.

Currently, Function Flow is still under heavy development. But we have made progresses, like easier programming than TBB and PPL, comparable performance to TBB.

###Why Function Flow?
Function Flow is highly motivated by our painful experiences while parallel programming. An example will be better than hundreds of words to explain it. As many other parallel libs are using Fibonacci as typical, we'd like to compare it fairly. The following code is from TBB's manual. As one can say, parallel programming with TBB is some "expert" thing, because it needs a lot of professional skills, or conventions.

    long ParallelFib( long n ) {
        long sum;
        FibTask& a = *new(task::allocate_root()) FibTask(n,&sum);
        task::spawn_root_and_wait(a);
        return sum;
    }

    class FibTask: public task {
    public:
        const long n;
        long* const sum;
        FibTask( long n_, long* sum_ ) :
            n(n_), sum(sum_)
        {}
        task* execute() {      // Overrides virtual function task::execute
        if( n<CutOff ) {
            *sum = SerialFib(n);
        } else {
            long x, y;
            FibTask& a = *new( allocate_child() ) FibTask(n-1,&x);
            FibTask& b = *new( allocate_child() ) FibTask(n-2,&y);
            // Set ref_count to 'two children plus one for the wait".
            set_ref_count(3);
            // Start b running.
            spawn( b );
            // Start a running and wait for all children (a and b).
            spawn_and_wait_for_all(a);
            // Do the sum
            *sum = x+y;
        }
        return NULL;
        }
    };
Please allow us to list some conventions to show whay parallel programming with TBB is a nightmare.
    1. new FibTask with "task::allocate_root"?
    2. spawn_root_and_wait(a), what's a root?
    3. derive from "task", and have to contains a "execute()" method?
    4. allocate_child(), uh, differences from allocate_root?
    5. set_ref_count(3). Gosh, of course, it's 3 although it's weired to write something the program "should know".

We believe that such conventions, or insights, or skills, are obstacles for popularizing parallel programming. So we decide to develop Function Flow to make everything is as intuitive as sequentially programming.

###An example of Function Flow
We'd like to show how Fibonacci is done in Function Flow, as shown in the following code.

    int fib(int n)
    {
        if(n <=2)
    		return 1;
        ff::para<int> a, b;
        a([&n]()->int{return fib(n - 1);});
        b([&n]()->int{return fib(n - 2);});
        return (a && b).then([](int x, int y){return x + y;});
    }
Function Flow loves lambda of C++11, so we use it heavily. We won't explain every corner of Function Flow here. But it's OK to know that *a* and *b* are tasks, and *a&&b* means waiting for *a* and *b*'s over to execute the lambda function in *then(...)*. Noticing that the parallel version of *Fibonacci* is not much longer than any sequential versions.

###Get it, and contact us
Function Flow is totally opensource and distributed under MIT lisence. Anyone can checkout the source code from GitHub. Please contact me via athrunarthur@gmail.com if you are interest or have any ideas of Function Flow. Enjoy it and have fun!

BTW, more details are located in docs/.
