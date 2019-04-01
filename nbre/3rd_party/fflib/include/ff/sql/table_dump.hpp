/***********************************************
  The MIT License (MIT)

  Copyright (c) 2012 Athrun Arthur <athrunarthur@gmail.com>

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in
  all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
  THE SOFTWARE.
 *************************************************/
template<typename T, typename T1, typename... TS>
static void recursive_dump_col_name(std::stringstream & ss){
  ss<<T::name<<", ";
  recursive_dump_col_name<T1, TS...>(ss);
}
template<typename T>
static void recursive_dump_col_name(std::stringstream & ss){
  ss<<T::name;
}

template<typename T, typename T1, typename... TS>
static void recursive_dump_col_creation(std::stringstream & ss){
  dump_col_creation<T>::dump(ss);
  ss<<", ";
  recursive_dump_col_creation<T1, TS...>(ss);
}
template<typename T>
static void recursive_dump_col_creation(std::stringstream & ss){
  dump_col_creation<T>::dump(ss);
  ss<<")";
}
template<typename engine_type, typename T, typename T1, typename... TS>
static void recursive_dump_for_index(engine_type * engine, std::stringstream & ss){
  if(std::is_base_of<index<typename T::type>, T>::value){
    ss.str(std::string());
    ss<<"create index "<<T::name<<"_index on "<<TM::table_name<<" ("<<T::name<<")";
    engine->eval_sql_string(ss.str());
    ss.str(std::string());
  }else{
  }
  recursive_dump_for_index<engine_type, T1, TS...>(engine, ss);
}
template<typename engine_type, typename T>
static void recursive_dump_for_index(engine_type * engine, std::stringstream & ss){
  if(std::is_base_of<index<typename T::type>, T>::value){
    ss<<"create index "<<T::name<<"_index on "<<TM::table_name<<" ("<<T::name<<")";
    ss.str(std::string());
    engine->eval_sql_string(ss.str());
    ss.str(std::string());
  }else{
  }
}

template<typename T, typename T1, typename... TS>
static void recurseive_dump_update_item_and_ignore_key(std::stringstream & ss){
  if(std::is_base_of<key<typename T::type>, T>::value){
  }else{
    ss<<T::name<<"=?, ";
  }
  recurseive_dump_update_item_and_ignore_key<T1, TS...>(ss);
}

template<typename T>
static void recurseive_dump_update_item_and_ignore_key(std::stringstream & ss){
  if(std::is_base_of<key<typename T::type>, T>::value){
    return ;
  }
  ss<<T::name<<"=?";
}

template<typename T, typename T1, typename... TS>
static void recurseive_dump_update_item_for_only_where(std::stringstream & ss){
  if(std::is_base_of<key<typename T::type>, T>::value){
    ss<<T::name<<"=?";
  }else{
    recurseive_dump_update_item_for_only_where<T1, TS...>(ss);
  }
}

template<typename T>
static void recurseive_dump_update_item_for_only_where(std::stringstream & ss){
  if(std::is_base_of<key<typename T::type>, T>::value){
    ss<<T::name<<"=?";
  }
}
