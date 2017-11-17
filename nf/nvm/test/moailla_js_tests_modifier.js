var native_getTestCaseResult = getTestCaseResult;

function getTestCaseResult2(expected, actual) {
    var ret = native_getTestCaseResult(expected, actual);
    if (!ret) {
        throw new Error(FAILED + 'expected is ' + expected + ', actual is ' + actual);
    }
    return ret;
}

getTestCaseResult = getTestCaseResult2;
