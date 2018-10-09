#include "bar.h"

void foo()
{
    bar();
}

int main(int argc, char *argv[])
{
    foo();
    return 0;
}
