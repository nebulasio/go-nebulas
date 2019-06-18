/*
 * Copyright (C) 2017 ETH Zurich, University of Bologna
 * All rights reserved.
 *
 *
 * Bug fixes and contributions will eventually be released under the
 * SolderPad open hardware license in the context of the PULP platform
 * (http://www.pulp-platform.org), under the copyright of ETH Zurich and the
 * University of Bologna.
 *
 * -------
 *
 * C++ wrapper for the SoftFloat library.
 * Defines the following classes:
 *  - float8
 *  - float16
 *  - float32
 *  - float64
 *  - extFloat80
 *  - float128
 * Includes overloaded operators for:
 *  - casts
 *  - arithmetic operators
 *  - relational operators
 *  - compound assignments
 *
 * -------
 *
 * Author: Stefan Mach
 * Email:  smach@iis.ee.ethz.ch
 * Date:   2017-05-19 09:51:34
 * Last Modified by:   Xuepeng Fan
 * Last Modified time: 2018-11-20
 */

#pragma once
#include "softfloat.h"
#include <string>
#include <stdexcept>

/*----------------------------------------------------------------------------
|  _______                   _       _
| |__   __|                 | |     | |
|    | | ___ _ __ ___  _ __ | | __ _| |_ ___  ___
|    | |/ _ \ '_ ` _ \| '_ \| |/ _` | __/ _ \/ __|
|    | |  __/ | | | | | |_) | | (_| | ||  __/\__ \
|    |_|\___|_| |_| |_| .__/|_|\__,_|\__\___||___/
|                     | |
|                     |_|
*----------------------------------------------------------------------------*/
/*----------------------------------------------------------------------------
| Function Templates
*----------------------------------------------------------------------------*/
template <typename From, typename To> static inline To softfloat_cast(const From &);
template <typename T> static inline T softfloat_roundToInt(const T &);
template <typename T> static inline T softfloat_add(const T &, const T &);
template <typename T> static inline T softfloat_sub(const T &, const T &);
template <typename T> static inline T softfloat_mul(const T &, const T &);
template <typename T> static inline T softfloat_mulAdd(const T &, const T &, const T &);
template <typename T> static inline T softfloat_div(const T &, const T &);
template <typename T> static inline T softfloat_rem(const T &, const T &);
template <typename T> static inline T softfloat_sqrt(const T &);
template <typename T> static inline bool softfloat_eq(const T &, const T &);
template <typename T> static inline bool softfloat_le(const T &, const T &);
template <typename T> static inline bool softfloat_lt(const T &, const T &);
template <typename T> static inline bool softfloat_isSignalingNaN(const T &);

/*----------------------------------------------------------------------------
| Class Template
*----------------------------------------------------------------------------*/
template <typename T> class softfloat;


/*----------------------------------------------------------------------------
|  ______                    _____        __ _       _ _   _
| |  ____|                  |  __ \      / _(_)     (_) | (_)
| | |__ _   _ _ __   ___    | |  | | ___| |_ _ _ __  _| |_ _  ___  _ __  ___
| |  __| | | | '_ \ / __|   | |  | |/ _ \  _| | '_ \| | __| |/ _ \| '_ \/ __|
| | |  | |_| | | | | (__ _  | |__| |  __/ | | | | | | | |_| | (_) | | | \__ \
| |_|   \__,_|_| |_|\___(_) |_____/ \___|_| |_|_| |_|_|\__|_|\___/|_| |_|___/
|
*----------------------------------------------------------------------------*/
/*----------------------------------------------------------------------------
|  __                                  __ _
| /   _  _ _|_ _     o __ _|_   ---   |_ |_)
| \__(_|_>  |__>     | | | |_         |  |
|
*----------------------------------------------------------------------------*/


/*----------------------------------------------------------------------------
| From ui32
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<uint32_t, float16_t>(const uint32_t &v) {
    return ui32_to_f16(v);
}
template<> float32_t softfloat_cast<std::string, float32_t>(const std::string & v){
  throw std::runtime_error("no impl here");
}

template <> float32_t softfloat_cast<uint32_t, float32_t>(const uint32_t &v) {
    return ui32_to_f32(v);
}

template <> float64_t softfloat_cast<uint32_t, float64_t>(const uint32_t &v) {
    return ui32_to_f64(v);
}

template <> extFloat80_t softfloat_cast<uint32_t, extFloat80_t>(const uint32_t &v) {
    extFloat80_t tmp;
    ui32_to_extF80M(v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<uint32_t, float128_t>(const uint32_t &v) {
    float128_t tmp;
    ui32_to_f128M(v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From ui64
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<uint64_t, float16_t>(const uint64_t &v) {
    return ui64_to_f16(v);
}

template <> float32_t softfloat_cast<uint64_t, float32_t>(const uint64_t &v) {
    return ui64_to_f32(v);
}

template <> float64_t softfloat_cast<uint64_t, float64_t>(const uint64_t &v) {
    return ui64_to_f64(v);
}

template <> extFloat80_t softfloat_cast<uint64_t, extFloat80_t>(const uint64_t &v) {
    extFloat80_t tmp;
    ui64_to_extF80M(v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<uint64_t, float128_t>(const uint64_t &v) {
    float128_t tmp;
    ui64_to_f128M(v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From i32
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<int32_t, float16_t>(const int32_t &v) {
    return i32_to_f16(v);
}

template <> float32_t softfloat_cast<int32_t, float32_t>(const int32_t &v) {
    return i32_to_f32(v);
}

template <> float64_t softfloat_cast<int32_t, float64_t>(const int32_t &v) {
    return i32_to_f64(v);
}

template <> extFloat80_t softfloat_cast<int32_t, extFloat80_t>(const int32_t &v) {
    extFloat80_t tmp;
    i32_to_extF80M(v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<int32_t, float128_t>(const int32_t &v) {
    float128_t tmp;
    i32_to_f128M(v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From i64
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<int64_t, float16_t>(const int64_t &v) {
    return i64_to_f16(v);
}

template <> float32_t softfloat_cast<int64_t, float32_t>(const int64_t &v) {
    return i64_to_f32(v);
}

template <> float64_t softfloat_cast<int64_t, float64_t>(const int64_t &v) {
    return i64_to_f64(v);
}

template <> extFloat80_t softfloat_cast<int64_t, extFloat80_t>(const int64_t &v) {
    extFloat80_t tmp;
    i64_to_extF80M(v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<int64_t, float128_t>(const int64_t &v) {
    float128_t tmp;
    i64_to_f128M(v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  __                __ _           __ _
| /   _  _ _|_ _    |_ |_)   ---   |_ |_)
| \__(_|_>  |__>    |  |           |  |
|
*----------------------------------------------------------------------------*/

/*----------------------------------------------------------------------------
| From float
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<float, float16_t>(const float &v) {
    float32_t result;
    const uint32_t *value = reinterpret_cast<const uint32_t *>(&v);
    result.v = *value;
    return f32_to_f16(result);
}

template <> float32_t softfloat_cast<float, float32_t>(const float &v) {
    float32_t result;
    const uint32_t *value = reinterpret_cast<const uint32_t *>(&v);
    result.v = *value;
    return result;
}

template <> float64_t softfloat_cast<float, float64_t>(const float &v) {
    float32_t result;
    const uint32_t *value = reinterpret_cast<const uint32_t *>(&v);
    result.v = *value;
    return f32_to_f64(result);
}

template <> extFloat80_t softfloat_cast<float, extFloat80_t>(const float &v) {
    float32_t result;
    const uint32_t *value = reinterpret_cast<const uint32_t *>(&v);
    result.v = *value;
    extFloat80_t tmp;
    f32_to_extF80M(result, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<float, float128_t>(const float &v) {
    float32_t result;
    const uint32_t *value = reinterpret_cast<const uint32_t *>(&v);
    result.v = *value;
    float128_t tmp;
    f32_to_f128M(result, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From double
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<double, float16_t>(const double &v) {
    float64_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v = *value;
    return f64_to_f16(result);
}

template <> float32_t softfloat_cast<double, float32_t>(const double &v) {
    float64_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v = *value;
    return f64_to_f32(result);
}

template <> float64_t softfloat_cast<double, float64_t>(const double &v) {
    float64_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v = *value;
    return result;
}

template <> extFloat80_t softfloat_cast<double, extFloat80_t>(const double &v) {
    float64_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v = *value;
    extFloat80_t tmp;
    f64_to_extF80M(result, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<double, float128_t>(const double &v) {
    float64_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v = *value;
    float128_t tmp;
    f64_to_f128M(result, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From long double
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<long double, float16_t>(const long double &v) {
    float128_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v[0] = value[0];
    result.v[1] = value[1];
    return f128M_to_f16(&result);
}

template <> float32_t softfloat_cast<long double, float32_t>(const long double &v) {
    float128_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v[0] = value[0];
    result.v[1] = value[1];
    return f128M_to_f32(&result);
}

template <> float64_t softfloat_cast<long double, float64_t>(const long double &v) {
    float128_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v[0] = value[0];
    result.v[1] = value[1];
    return f128M_to_f64(&result);
}

template <> extFloat80_t softfloat_cast<long double, extFloat80_t>(const long double &v) {
    float128_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v[0] = value[0];
    result.v[1] = value[1];
    extFloat80_t tmp;
    f128M_to_extF80M(&result, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<long double, float128_t>(const long double &v) {
    float128_t result;
    const uint64_t *value = reinterpret_cast<const uint64_t *>(&v);
    result.v[0] = value[0];
    result.v[1] = value[1];
    return result;
}


/*----------------------------------------------------------------------------
| From f16
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<float16_t, float16_t>(const float16_t &v) {
    return v;
}

template <> float32_t softfloat_cast<float16_t, float32_t>(const float16_t &v) {
    return f16_to_f32(v);
}

template <> float64_t softfloat_cast<float16_t, float64_t>(const float16_t &v) {
    return f16_to_f64(v);
}

template <> extFloat80_t softfloat_cast<float16_t, extFloat80_t>(const float16_t &v) {
    extFloat80_t tmp;
    f16_to_extF80M(v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<float16_t, float128_t>(const float16_t &v) {
    float128_t tmp;
    f16_to_f128M(v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From f32
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<float32_t, float16_t>(const float32_t &v) {
    return f32_to_f16(v);
}

template <> float32_t softfloat_cast<float32_t, float32_t>(const float32_t &v) {
    return v;
}

template <> float64_t softfloat_cast<float32_t, float64_t>(const float32_t &v) {
    return f32_to_f64(v);
}

template <> extFloat80_t softfloat_cast<float32_t, extFloat80_t>(const float32_t &v) {
    extFloat80_t tmp;
    f32_to_extF80M(v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<float32_t, float128_t>(const float32_t &v) {
    float128_t tmp;
    f32_to_f128M(v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From f64
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<float64_t, float16_t>(const float64_t &v) {
    return f64_to_f16(v);
}

template <> float32_t softfloat_cast<float64_t, float32_t>(const float64_t &v) {
    return f64_to_f32(v);
}

template <> float64_t softfloat_cast<float64_t, float64_t>(const float64_t &v) {
    return v;
}

template <> extFloat80_t softfloat_cast<float64_t, extFloat80_t>(const float64_t &v) {
    extFloat80_t tmp;
    f64_to_extF80M(v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<float64_t, float128_t>(const float64_t &v) {
    float128_t tmp;
    f64_to_f128M(v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From extF80
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<extFloat80_t, float16_t>(const extFloat80_t &v) {
    return extF80M_to_f16(&v);
}

template <> float32_t softfloat_cast<extFloat80_t, float32_t>(const extFloat80_t &v) {
    return extF80M_to_f32(&v);
}

template <> float64_t softfloat_cast<extFloat80_t, float64_t>(const extFloat80_t &v) {
    return extF80M_to_f64(&v);
}

template <> extFloat80_t softfloat_cast<extFloat80_t, extFloat80_t>(const extFloat80_t &v) {
    return v;
}

template <> float128_t softfloat_cast<extFloat80_t, float128_t>(const extFloat80_t &v) {
    float128_t tmp;
    extF80M_to_f128M(&v, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
| From f128
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_cast<float128_t, float16_t>(const float128_t &v) {
    return f128M_to_f16(&v);
}

template <> float32_t softfloat_cast<float128_t, float32_t>(const float128_t &v) {
    return f128M_to_f32(&v);
}

template <> float64_t softfloat_cast<float128_t, float64_t>(const float128_t &v) {
    return f128M_to_f64(&v);
}

template <> extFloat80_t softfloat_cast<float128_t, extFloat80_t>(const float128_t &v) {
    extFloat80_t tmp;
    f128M_to_extF80M(&v, &tmp);
    return tmp;
}

template <> float128_t softfloat_cast<float128_t, float128_t>(const float128_t &v) {
    return v;
}

/*----------------------------------------------------------------------------
|  __                __ _
| /   _  _ _|_ _    |_ |_)   ---    o __ _|_
| \__(_|_>  |__>    |  |            | | | |_
|
| The template specializations here use round towards 0 as their default mode!
*----------------------------------------------------------------------------*/

/*----------------------------------------------------------------------------
| From f16
*----------------------------------------------------------------------------*/
template <> uint32_t softfloat_cast<float16_t, uint32_t>(const float16_t &v) {
    return f16_to_ui32_r_minMag(v, true);
}

template <> uint64_t softfloat_cast<float16_t, uint64_t>(const float16_t &v) {
    return f16_to_ui64_r_minMag(v, true);
}

template <> int32_t softfloat_cast<float16_t, int32_t>(const float16_t &v) {
    return f16_to_i32_r_minMag(v, true);
}

template <> int64_t softfloat_cast<float16_t, int64_t>(const float16_t &v) {
    return f16_to_i64_r_minMag(v, true);
}

/*----------------------------------------------------------------------------
| From f32
*----------------------------------------------------------------------------*/
template <> uint32_t softfloat_cast<float32_t, uint32_t>(const float32_t &v) {
    return f32_to_ui32_r_minMag(v, true);
}

template <> uint64_t softfloat_cast<float32_t, uint64_t>(const float32_t &v) {
    return f32_to_ui64_r_minMag(v, true);
}

template <> int32_t softfloat_cast<float32_t, int32_t>(const float32_t &v) {
    return f32_to_i32_r_minMag(v, true);
}

template <> int64_t softfloat_cast<float32_t, int64_t>(const float32_t &v) {
    return f32_to_i64_r_minMag(v, true);
}

/*----------------------------------------------------------------------------
| From f64
*----------------------------------------------------------------------------*/
template <> uint32_t softfloat_cast<float64_t, uint32_t>(const float64_t &v) {
    return f64_to_ui32_r_minMag(v, true);
}

template <> uint64_t softfloat_cast<float64_t, uint64_t>(const float64_t &v) {
    return f64_to_ui64_r_minMag(v, true);
}

template <> int32_t softfloat_cast<float64_t, int32_t>(const float64_t &v) {
    return f64_to_i32_r_minMag(v, true);
}

template <> int64_t softfloat_cast<float64_t, int64_t>(const float64_t &v) {
    return f64_to_i64_r_minMag(v, true);
}

/*----------------------------------------------------------------------------
| From ext80
*----------------------------------------------------------------------------*/
template <> uint32_t softfloat_cast<extFloat80_t, uint32_t>(const extFloat80_t &v) {
    return extF80M_to_ui32_r_minMag(&v, true);
}

template <> uint64_t softfloat_cast<extFloat80_t, uint64_t>(const extFloat80_t &v) {
    return extF80M_to_ui64_r_minMag(&v, true);
}

template <> int32_t softfloat_cast<extFloat80_t, int32_t>(const extFloat80_t &v) {
    return extF80M_to_i32_r_minMag(&v, true);
}

template <> int64_t softfloat_cast<extFloat80_t, int64_t>(const extFloat80_t &v) {
    return extF80M_to_i64_r_minMag(&v, true);
}

/*----------------------------------------------------------------------------
| From f128
*----------------------------------------------------------------------------*/
template <> uint32_t softfloat_cast<float128_t, uint32_t>(const float128_t &v) {
    return f128M_to_ui32_r_minMag(&v, true);
}

template <> uint64_t softfloat_cast<float128_t, uint64_t>(const float128_t &v) {
    return f128M_to_ui64_r_minMag(&v, true);
}

template <> int32_t softfloat_cast<float128_t, int32_t>(const float128_t &v) {
    return f128M_to_i32_r_minMag(&v, true);
}

template <> int64_t softfloat_cast<float128_t, int64_t>(const float128_t &v) {
    return f128M_to_i64_r_minMag(&v, true);
}

/*----------------------------------------------------------------------------
|  _
| |_) _    __  _|   _|_ _     o __ _|_
| | \(_)|_|| |(_|    |_(_)    | | | |_
|
| These template specializations use the global rounding mode for rounding!
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_roundToInt(const float16_t &v) {
    return f16_roundToInt(v, softfloat_roundingMode, true);
}

template <> float32_t softfloat_roundToInt(const float32_t &v) {
    return f32_roundToInt(v, softfloat_roundingMode, true);
}

template <> float64_t softfloat_roundToInt(const float64_t &v) {
    return f64_roundToInt(v, softfloat_roundingMode, true);
}

template <> extFloat80_t softfloat_roundToInt(const extFloat80_t &v) {
    extFloat80_t tmp;
    extF80M_roundToInt(&v, softfloat_roundingMode, true, &tmp);
    return tmp;
}

template <> float128_t softfloat_roundToInt(const float128_t &v) {
    float128_t tmp;
    f128M_roundToInt(&v, softfloat_roundingMode, true, &tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  _
| |_| _| _| o _|_ o  _ __
| | |(_|(_| |  |_ | (_)| |
|
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_add(const float16_t &a, const float16_t &b) {
    return f16_add(a,b);
}

template <> float32_t softfloat_add(const float32_t &a, const float32_t &b) {
    return f32_add(a,b);
}

template <> float64_t softfloat_add(const float64_t &a, const float64_t &b) {
    return f64_add(a,b);
}

template <> extFloat80_t softfloat_add(const extFloat80_t &a, const extFloat80_t &b) {
    extFloat80_t tmp;
    extF80M_add(&a,&b,&tmp);
    return tmp;
}

template <> float128_t softfloat_add(const float128_t &a, const float128_t &b) {
    float128_t tmp;
    f128M_add(&a,&b,&tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  __
| (_    |_ _|_ __ _  _ _|_ o  _ __
| __)|_||_) |_ | (_|(_  |_ | (_)| |
|
*----------------------------------------------------------------------------*/

template <> float16_t softfloat_sub(const float16_t &a, const float16_t &b) {
    return f16_sub(a,b);
}

template <> float32_t softfloat_sub(const float32_t &a, const float32_t &b) {
    return f32_sub(a,b);
}

template <> float64_t softfloat_sub(const float64_t &a, const float64_t &b) {
    return f64_sub(a,b);
}

template <> extFloat80_t softfloat_sub(const extFloat80_t &a, const extFloat80_t &b) {
    extFloat80_t tmp;
    extF80M_sub(&a,&b,&tmp);
    return tmp;
}

template <> float128_t softfloat_sub(const float128_t &a, const float128_t &b) {
    float128_t tmp;
    f128M_sub(&a,&b,&tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|                 _
| |V|    | _|_ o |_) |  o  _  _ _|_ o  _ __
| | ||_| |  |_ | |   |  | (_ (_| |_ | (_)| |
|
*----------------------------------------------------------------------------*/
template <> float16_t softfloat_mul(const float16_t &a, const float16_t &b) {
    return f16_mul(a,b);
}

template <> float32_t softfloat_mul(const float32_t &a, const float32_t &b) {
    return f32_mul(a,b);
}

template <> float64_t softfloat_mul(const float64_t &a, const float64_t &b) {
    return f64_mul(a,b);
}

template <> extFloat80_t softfloat_mul(const extFloat80_t &a, const extFloat80_t &b) {
    extFloat80_t tmp;
    extF80M_mul(&a,&b,&tmp);
    return tmp;
}

template <> float128_t softfloat_mul(const float128_t &a, const float128_t &b) {
    float128_t tmp;
    f128M_mul(&a,&b,&tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  __                               _           _
| |_     _  _  _|   |V|    | _|_ o |_) |  \/---|_| _| _|
| |  |_|_> (/_(_|   | ||_| |  |_ | |   |  /    | |(_|(_|
|
*----------------------------------------------------------------------------*/
template <> float16_t softfloat_mulAdd(const float16_t &a, const float16_t &b, const float16_t &c) {
    return f16_mulAdd(a,b,c);
}

template <> float32_t softfloat_mulAdd(const float32_t &a, const float32_t &b, const float32_t &c) {
    return f32_mulAdd(a,b,c);
}

template <> float64_t softfloat_mulAdd(const float64_t &a, const float64_t &b, const float64_t &c) {
    return f64_mulAdd(a,b,c);
}

template <> extFloat80_t softfloat_mulAdd(const extFloat80_t &a, const extFloat80_t &b, const extFloat80_t &c) {
    return softfloat_add(softfloat_mul(a,b),c); /* Compounded */
}

template <> float128_t softfloat_mulAdd(const float128_t &a, const float128_t &b, const float128_t &c) {
    float128_t tmp;
    f128M_mulAdd(&a,&b,&c,&tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  _
| | \ o     o  _  o  _ __
| |_/ | \_/ | _>  | (_)| |
|
*----------------------------------------------------------------------------*/
template <> float16_t softfloat_div(const float16_t &a, const float16_t &b) {
    return f16_div(a,b);
}

template <> float32_t softfloat_div(const float32_t &a, const float32_t &b) {
    return f32_div(a,b);
}

template <> float64_t softfloat_div(const float64_t &a, const float64_t &b) {
    return f64_div(a,b);
}

template <> extFloat80_t softfloat_div(const extFloat80_t &a, const extFloat80_t &b) {
    extFloat80_t tmp;
    extF80M_div(&a,&b,&tmp);
    return tmp;
}

template <> float128_t softfloat_div(const float128_t &a, const float128_t &b) {
    float128_t tmp;
    f128M_div(&a,&b,&tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  _
| |_) _ __  _  o __  _| _  __
| | \(/_|||(_| | | |(_|(/_ |
|
*----------------------------------------------------------------------------*/
template <> float16_t softfloat_rem(const float16_t &a, const float16_t &b) {
    return f16_rem(a,b);
}

template <> float32_t softfloat_rem(const float32_t &a, const float32_t &b) {
    return f32_rem(a,b);
}

template <> float64_t softfloat_rem(const float64_t &a, const float64_t &b) {
    return f64_rem(a,b);
}

template <> extFloat80_t softfloat_rem(const extFloat80_t &a, const extFloat80_t &b) {
    extFloat80_t tmp;
    extF80M_rem(&a,&b,&tmp);
    return tmp;
}

template <> float128_t softfloat_rem(const float128_t &a, const float128_t &b) {
    float128_t tmp;
    f128M_rem(&a,&b,&tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  __ _                 _
| (_ (_|    _  __ _    |_) _  _ _|_
| __)  ||_|(_| | (/_   | \(_)(_) |_
|
*----------------------------------------------------------------------------*/
template <> float16_t softfloat_sqrt(const float16_t &a) {
    return f16_sqrt(a);
}

template <> float32_t softfloat_sqrt(const float32_t &a) {
    return f32_sqrt(a);
}

template <> float64_t softfloat_sqrt(const float64_t &a) {
    return f64_sqrt(a);
}

template <> extFloat80_t softfloat_sqrt(const extFloat80_t &a) {
    extFloat80_t tmp;
    extF80M_sqrt(&a,&tmp);
    return tmp;
}

template <> float128_t softfloat_sqrt(const float128_t &a) {
    float128_t tmp;
    f128M_sqrt(&a,&tmp);
    return tmp;
}

/*----------------------------------------------------------------------------
|  __ _
| |_ (_|    _  |  o _|_ \/
| |__  ||_|(_| |  |  |_ /
|
| These template specializations use signaling comparisons!
*----------------------------------------------------------------------------*/
template <> bool softfloat_eq(const float16_t &a, const float16_t &b) {
    return f16_eq_signaling(a,b);
}

template <> bool softfloat_eq(const float32_t &a, const float32_t &b) {
    return f32_eq_signaling(a,b);
}

template <> bool softfloat_eq(const float64_t &a, const float64_t &b) {
    return f64_eq_signaling(a,b);
}

template <> bool softfloat_eq(const extFloat80_t &a, const extFloat80_t &b) {
     return extF80M_eq_signaling(&a,&b);
}

template <> bool softfloat_eq(const float128_t &a, const float128_t &b) {
     return f128M_eq_signaling(&a,&b);
}

/*----------------------------------------------------------------------------
|                ___             _        __ _             ___
| |   _  _  _ --- | |_  _ __ ---/ \ __---|_ (_|    _  | --- |  _
| |__(/__> _>     | | |(_|| |   \_/ |    |__  ||_|(_| |     | (_)
|
| These template specializations use signaling comparisons!
*----------------------------------------------------------------------------*/
template <> bool softfloat_le(const float16_t &a, const float16_t &b) {
    return f16_le(a,b);
}

template <> bool softfloat_le(const float32_t &a, const float32_t &b) {
    return f32_le(a,b);
}

template <> bool softfloat_le(const float64_t &a, const float64_t &b) {
    return f64_le(a,b);
}

template <> bool softfloat_le(const extFloat80_t &a, const extFloat80_t &b) {
     return extF80M_le(&a,&b);
}

template <> bool softfloat_le(const float128_t &a, const float128_t &b) {
     return f128M_le(&a,&b);
}

/*----------------------------------------------------------------------------
|                ___
| |   _  _  _ --- | |_  _ __
| |__(/__> _>     | | |(_|| |
|
| These template specializations use signaling comparisons!
*----------------------------------------------------------------------------*/
template <> bool softfloat_lt(const float16_t &a, const float16_t &b) {
    return f16_lt(a,b);
}

template <> bool softfloat_lt(const float32_t &a, const float32_t &b) {
    return f32_lt(a,b);
}

template <> bool softfloat_lt(const float64_t &a, const float64_t &b) {
    return f64_lt(a,b);
}

template <> bool softfloat_lt(const extFloat80_t &a, const extFloat80_t &b) {
     return extF80M_lt(&a,&b);
}

template <> bool softfloat_lt(const float128_t &a, const float128_t &b) {
     return f128M_le(&a,&b);
}

/*----------------------------------------------------------------------------
| ___       __    _                 _
|  |  _    (_  o (_|__  _  |  o __ (_|   |\| _ |\|
| _|__>    __) | __|| |(_| |  | | |__|   | |(_|| |
|
*----------------------------------------------------------------------------*/
template <> bool softfloat_isSignalingNaN(const float16_t &a) {
    return f16_isSignalingNaN(a);
}

template <> bool softfloat_isSignalingNaN(const float32_t &a) {
    return f32_isSignalingNaN(a);
}

template <> bool softfloat_isSignalingNaN(const float64_t &a) {
    return f64_isSignalingNaN(a);
}

template <> bool softfloat_isSignalingNaN(const extFloat80_t &a) {
     return extF80M_isSignalingNaN(&a);
}

template <> bool softfloat_isSignalingNaN(const float128_t &a) {
     return f128M_isSignalingNaN(&a);
}


/*----------------------------------------------------------------------------
|   _____ _                 _____        __ _       _ _   _
|  / ____| |               |  __ \      / _(_)     (_) | (_)
| | |    | | __ _ ___ ___  | |  | | ___| |_ _ _ __  _| |_ _  ___  _ __
| | |    | |/ _` / __/ __| | |  | |/ _ \  _| | '_ \| | __| |/ _ \| '_ \
| | |____| | (_| \__ \__ \ | |__| |  __/ | | | | | | | |_| | (_) | | | |
|  \_____|_|\__,_|___/___/ |_____/ \___|_| |_|_| |_|_|\__|_|\___/|_| |_|
|
*----------------------------------------------------------------------------*/

#include <iostream>

template<typename TU>
struct redirect_is_signed{
  constexpr static bool value = std::is_same<TU, long>::value || std::is_same<TU, long long>::value
    || std::is_same<TU, short>::value || std::is_same<TU, int>::value;
};

template<typename TU>
struct redirect_is_unsigned{
  constexpr static bool value = std::is_same<TU, unsigned long>::value || std::is_same<TU, unsigned long long>::value
    || std::is_same<TU, unsigned short>::value || std::is_same<TU, unsigned int>::value;
};

template<typename TU, bool needs_redirect = redirect_is_signed<TU>::value || redirect_is_unsigned<TU>::value >
struct infer_type_helper{};
template<typename TU>
struct infer_type_helper<TU, false>{
  typedef TU type;
};

template <size_t L> struct sized_int_type{};
template <> struct sized_int_type<2>{
  typedef int16_t type;
};
template <> struct sized_int_type<4>{
  typedef int32_t type;
};
template <> struct sized_int_type<8>{
  typedef int64_t type;
};

template <size_t L> struct sized_uint_type{};
template <> struct sized_uint_type<2>{
  typedef uint16_t type;
};
template <> struct sized_uint_type<4>{
  typedef uint32_t type;
};
template <> struct sized_uint_type<8>{
  typedef uint64_t type;
};

template <typename TU>
struct infer_type_helper<TU, true>{
  typedef typename std::conditional_t<redirect_is_signed<TU>::value,
          typename sized_int_type<sizeof(TU)>::type,
          typename sized_uint_type<sizeof(TU)>::type> type ;
};
template <typename T> class softfloat {
  public:
    typedef T value_type;
protected:
    T v;

public:
    // Empty constructor --> initialize to positive zero.
    inline softfloat () {
        v = softfloat_cast<uint32_t,T>(0);
    }

    // Constructor from wrapped type T
    inline softfloat (const T &v) : v(v) {}

    // Constructor from softfloat types
    template <typename U> inline softfloat (const softfloat<U> &w) {
        v = softfloat_cast<U,T>(w);
    }

    // Constructor from castable type
    template <typename U> inline softfloat (const U &w) {
        typedef typename infer_type_helper<U>::type TU;
        v = softfloat_cast<TU,T>(w);
    }

    inline softfloat<T> round_to_int(){
      return softfloat(softfloat_roundToInt(v));
    }
    inline softfloat<T> integer_val(){
      auto prev = softfloat_roundingMode;
      softfloat_roundingMode = softfloat_round_min;
      auto t = softfloat(softfloat_roundToInt(v));
      softfloat_roundingMode = prev;
      return t;
    }
    inline softfloat<T> decimal_val(){
      return *this - integer_val();
    }

    T value(){return v;}



    /*------------------------------------------------------------------------
    | OPERATOR OVERLOADS: CASTS
    *------------------------------------------------------------------------*/

    // Cast to the wrapped types --> can implicitly use softfloat object with
    // <float>_ functions from C package.
    inline operator T() const {
        return v;
    }

    inline explicit operator float() const {
      uint32_t temp = softfloat_cast<T, float32_t>(v).v;
      float *result = reinterpret_cast<float *>(&temp);
      return *result;
    }

    inline explicit operator double() const {
      uint64_t temp = softfloat_cast<T, float64_t>(v).v;
      double *result = reinterpret_cast<double *>(&temp);
      return *result;
    }

    inline explicit operator long double() const {
      extFloat80_t temp = softfloat_cast<T, extFloat80_t >(v);
      long double *result = reinterpret_cast<long double *>(&temp);
      return *result;
    }


    /*------------------------------------------------------------------------
    | OPERATOR OVERLOADS: Arithmetics
    *------------------------------------------------------------------------*/


    /* UNARY MINUS (-) */
    inline softfloat operator-() const
    {
        return softfloat_sub(softfloat_cast<uint32_t,T>(0), v);
    }

    /* UNARY PLUS (+) */
    inline softfloat operator+() const
    {
        return softfloat(*this);
    }

    /* ADD (+) */
    friend inline softfloat operator+(const softfloat &a, const softfloat &b)
    {
        return softfloat_add(a.v,b.v);
    }

    /* SUBTRACT (-) */
    friend inline softfloat operator-(const softfloat &a, const softfloat &b)
    {
        return softfloat_sub(a.v,b.v);
    }

    /* MULTIPLY (*) */
    friend inline softfloat operator*(const softfloat &a, const softfloat &b)
    {
        return softfloat_mul(a.v,b.v);
    }

    /* DIVIDE (/) */
    friend inline softfloat operator/(const softfloat &a, const softfloat &b)
    {
        return softfloat_div(a.v,b.v);
    }

    /*------------------------------------------------------------------------
    | OPERATOR OVERLOADS: Relational operators
    *------------------------------------------------------------------------*/

    /* EQUALITY (==) */
    inline bool operator==(const softfloat &b) const {
        return softfloat_eq(v,b.v);
    }

    /* INEQUALITY (!=) */
    inline bool operator!=(const softfloat &b) const {
        return !(softfloat_eq(v,b.v));
    }

    /* GREATER-THAN (>) */
    inline bool operator>(const softfloat &b) const {
        return !(softfloat_le(v,b.v));
    }

    /* LESS-THAN (<) */
    inline bool operator<(const softfloat &b) const {
        return softfloat_lt(v,b.v);
    }

    /* GREATER-THAN-OR-EQUAL-TO (>=) */
    inline bool operator>=(const softfloat &b) const {
        return !(softfloat_lt(v,b.v));
    }

    /* LESS-THAN-OR-EQUAL-TO (<=) */
    inline bool operator<=(const softfloat &b) const {
        return softfloat_le(v,b.v);
    }

    /*------------------------------------------------------------------------
    | OPERATOR OVERLOADS: Compound assignment operators (no bitwise ops)
    *------------------------------------------------------------------------*/
    inline softfloat &operator+=(const softfloat &b) {
        return *this = *this + b;
    }

    inline softfloat &operator-=(const softfloat &b) {
        return *this = *this - b;
    }

    inline softfloat &operator*=(const softfloat &b) {
        return *this = *this * b;
    }

    inline softfloat &operator/=(const softfloat &b) {
        return *this = *this / b;
    }

    /*------------------------------------------------------------------------
    | OPERATOR OVERLOADS: IO streams operators
    *------------------------------------------------------------------------*/
    friend std::ostream& operator<<(std::ostream& os, const softfloat& obj)
    {
      static_assert(sizeof(long double) == 16, "long double is too short");
      auto t = softfloat_cast<T, float64_t>(obj.v);
      os << *(double*)(&t);
      return os;
    }

    friend std::istream& operator>>(std::istream& is, softfloat& obj)
    {
      static_assert(sizeof(long double) == 16, "long double is too short");
      float64_t t;
      is >> *(double*)(&t);
      obj.v = softfloat_cast<float64_t, T>(t);
      return is;
    }
};



/*----------------------------------------------------------------------------
|  _______                   _       __
| |__   __|                 | |     / _|
|    | |_   _ _ __   ___  __| | ___| |_ ___
|    | | | | | '_ \ / _ \/ _` |/ _ \  _/ __|
|    | | |_| | |_) |  __/ (_| |  __/ | \__ \
|    |_|\__, | .__/ \___|\__,_|\___|_| |___/
|        __/ | |
|       |___/|_|
*----------------------------------------------------------------------------*/

typedef softfloat<float16_t>    float16;
typedef softfloat<float32_t>    float32;
typedef softfloat<float64_t>    float64;
typedef softfloat<extFloat80_t> extFloat80;
typedef softfloat<float128_t>   float128;

